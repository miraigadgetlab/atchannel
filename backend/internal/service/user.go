package service

import (
	"context"
	"errors"
	"slices"
	"time"

	"atchannel-backend/internal/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrInvalidCredentials = errors.New("invalid name or password")
	ErrNameTaken          = errors.New("name is already taken")
	ErrUserNotFound       = errors.New("user not found")
	ErrCurrentPassword    = errors.New("current password is incorrect")
	ErrAccountBanned      = errors.New("account is banned")
)

// UpdateProfileInput carries only the fields the client wants to change;
// a nil field is left untouched.
type UpdateProfileInput struct {
	Name      *string
	AboutMe   *string
	AvatarURL *string
	Note      *string
}

// AllowedRoles is the closed set of roles a channeler can hold. "moderator"
// sits between user and admin: it can edit/delete any content, but cannot
// manage accounts, roles or see stats.
var AllowedRoles = []string{"user", "moderator", "admin"}

type UserService struct {
	db *gorm.DB
}

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{db: db}
}

func (s *UserService) Create(ctx context.Context, name, password, note string) (*models.User, error) {
	var count int64

	if err := s.db.WithContext(ctx).Model(&models.User{}).Where("lower(name) = lower(?)", name).Count(&count).Error; err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, ErrNameTaken
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := models.User{
		Name:     name,
		Password: string(hashed),
		Roles:    []string{"user"},
		Note:     note,
	}

	if err := s.db.WithContext(ctx).Create(&user).Error; err != nil {
		// The pre-insert check cannot win a race against a concurrent
		// signup; catching the duplicate key closes the gap so the loser
		// gets the same 409 instead of an opaque 500.
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, ErrNameTaken
		}
		return nil, err
	}

	return &user, nil
}

func (s *UserService) Authenticate(ctx context.Context, name, password string) (*models.User, error) {
	var user models.User

	err := s.db.WithContext(ctx).Where("lower(name) = lower(?)", name).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	// A banned account must not get in with the (possibly leaked) correct
	// password either.
	if user.BannedAt != nil {
		return nil, ErrAccountBanned
	}

	return &user, nil
}

// SetBanned suspends or reinstates an account. Banning keeps the data (posts,
// comments) for the audit trail and sets BannedAt so every fresh login and new
// token is refused.
func (s *UserService) SetBanned(ctx context.Context, userID uint, banned bool) error {
	var user models.User
	if err := s.db.WithContext(ctx).First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return err
	}

	var value any
	if banned {
		value = time.Now()
	} else {
		value = nil
	}
	return s.db.WithContext(ctx).Model(&user).Update("banned_at", value).Error
}

// ValidRoles reports whether roles is non-empty and only contains allowed values.
func ValidRoles(roles []string) bool {
	if len(roles) == 0 {
		return false
	}

	seen := make(map[string]bool, len(roles))
	for _, role := range roles {
		if !slices.Contains(AllowedRoles, role) || seen[role] {
			return false
		}
		seen[role] = true
	}

	return true
}

func (s *UserService) List(ctx context.Context, limit, offset int) ([]models.User, int64, error) {
	query := s.db.WithContext(ctx)

	var total int64
	if err := query.Model(&models.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	var users []models.User
	if err := query.Order("id").Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (s *UserService) GetByID(ctx context.Context, userID uint) (*models.User, error) {
	var user models.User

	if err := s.db.WithContext(ctx).First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

// GetByName looks up an account by its display name (case-insensitively).
func (s *UserService) GetByName(ctx context.Context, name string) (*models.User, error) {
	var user models.User

	if err := s.db.WithContext(ctx).Where("lower(name) = lower(?)", name).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

// GetPublicProfile returns an account together with how much it has posted and
// commented, for the public profile page.
func (s *UserService) GetPublicProfile(ctx context.Context, userID uint) (*models.User, int64, int64, error) {
	user, err := s.GetByID(ctx, userID)
	if err != nil {
		return nil, 0, 0, err
	}

	var posts, comments int64
	if err := s.db.WithContext(ctx).Model(&models.Post{}).Where("user_id = ?", userID).Count(&posts).Error; err != nil {
		return nil, 0, 0, err
	}
	if err := s.db.WithContext(ctx).Model(&models.Comment{}).Where("user_id = ?", userID).Count(&comments).Error; err != nil {
		return nil, 0, 0, err
	}

	return user, posts, comments, nil
}

// GetPublicProfileByName returns an account together with how much it has posted and commented,
// for the public profile page.
func (s *UserService) GetPublicProfileByName(ctx context.Context, name string) (*models.User, int64, int64, error) {
	user, err := s.GetByName(ctx, name)
	if err != nil {
		return nil, 0, 0, err
	}
	return s.GetPublicProfile(ctx, user.ID)
}

// ChannelerRank is one row of the public channeler leaderboard: an account
// plus how many posts it has written.
type ChannelerRank struct {
	User      models.User
	PostCount int64
}

// Channelers returns accounts ordered by how many posts they have written,
// most active first. Only accounts that have posted at least once and are not
// banned appear — a leaderboard of people who never wrote anything (or were
// suspended) is noise. The total matches the same filter so the client can
// page through it.
func (s *UserService) Channelers(ctx context.Context, limit, offset int) ([]ChannelerRank, int64, error) {
	const postCountExpr = "(SELECT COUNT(*) FROM posts WHERE posts.user_id = users.id)"
	where := "users.banned_at IS NULL AND " + postCountExpr + " > 0"

	var total int64
	if err := s.db.WithContext(ctx).Model(&models.User{}).Where(where).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	type row struct {
		models.User
		PostCount int64 `gorm:"column:post_count"`
	}

	query := s.db.WithContext(ctx).
		Model(&models.User{}).
		Where(where).
		Select("users.*, " + postCountExpr + " AS post_count").
		Order("post_count DESC, users.id ASC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	var rows []row
	if err := query.Scan(&rows).Error; err != nil {
		return nil, 0, err
	}

	ranks := make([]ChannelerRank, len(rows))
	for i, r := range rows {
		ranks[i] = ChannelerRank{User: r.User, PostCount: r.PostCount}
	}

	return ranks, total, nil
}

// UpdateProfile applies only the provided fields; nil fields stay untouched.
// The name is the account's login identity, so changing it requires the
// current password: a borrowed session must not be able to quietly move the
// account out from under its owner. Name uniqueness is re-checked against
// every other account (case-insensitively).
func (s *UserService) UpdateProfile(ctx context.Context, userID uint, currentPassword string, in UpdateProfileInput) (*models.User, error) {
	user, err := s.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	nameChanging := in.Name != nil && *in.Name != user.Name
	if nameChanging {
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(currentPassword)); err != nil {
			return nil, ErrCurrentPassword
		}

		var count int64
		if err := s.db.WithContext(ctx).Model(&models.User{}).
			Where("lower(name) = lower(?) AND id <> ?", *in.Name, userID).Count(&count).Error; err != nil {
			return nil, err
		}
		if count > 0 {
			return nil, ErrNameTaken
		}
		user.Name = *in.Name
	}

	if in.AboutMe != nil {
		user.AboutMe = *in.AboutMe
	}
	if in.AvatarURL != nil {
		user.AvatarUrl = *in.AvatarURL
	}
	if in.Note != nil {
		user.Note = *in.Note
	}

	if err := s.db.WithContext(ctx).Save(user).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, ErrNameTaken
		}
		return nil, err
	}

	return user, nil
}

// UpdateAvatar stores the served path of a freshly uploaded avatar.
func (s *UserService) UpdateAvatar(ctx context.Context, userID uint, avatarURL string) (*models.User, error) {
	user, err := s.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if err := s.db.WithContext(ctx).Model(user).Update("avatar_url", avatarURL).Error; err != nil {
		return nil, err
	}
	user.AvatarUrl = avatarURL

	return user, nil
}

// ChangePassword verifies the current password before storing the new hash.
func (s *UserService) ChangePassword(ctx context.Context, userID uint, currentPassword, newPassword string) error {
	user, err := s.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(currentPassword)); err != nil {
		return ErrCurrentPassword
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return s.db.WithContext(ctx).Model(user).Update("password", string(hashed)).Error
}

func (s *UserService) SetRoles(ctx context.Context, userID uint, roles []string) (*models.User, error) {
	var user models.User

	if err := s.db.WithContext(ctx).First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	user.Roles = roles

	if err := s.db.WithContext(ctx).Save(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

// Delete removes a user together with everything they created: their comments,
// their posts (and the comments on those posts), and the channels they opened
// — a channel is owned by exactly one account, so leaving it behind would
// strand an ownerless channel that only an admin could ever touch.
func (s *UserService) Delete(ctx context.Context, userID uint) error {
	var user models.User

	if err := s.db.WithContext(ctx).First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return err
	}

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("DELETE FROM refresh_tokens WHERE user_id = ?", userID).Error; err != nil {
			return err
		}
		// Recovery codes only make sense for an account that exists.
		if err := tx.Exec("DELETE FROM recovery_codes WHERE user_id = ?", userID).Error; err != nil {
			return err
		}
		// notifications.actor_id references users(id), so these have to go
		// before the account row itself.
		if err := tx.Exec("DELETE FROM notifications WHERE user_id = ? OR actor_id = ?", userID, userID).Error; err != nil {
			return err
		}
		if err := tx.Exec("DELETE FROM post_likes WHERE user_id = ?", userID).Error; err != nil {
			return err
		}
		if err := tx.Exec("DELETE FROM post_likes WHERE post_id IN (SELECT id FROM posts WHERE user_id = ?)", userID).Error; err != nil {
			return err
		}
		if err := tx.Exec("DELETE FROM post_likes WHERE post_id IN (SELECT id FROM posts WHERE channel_id IN (SELECT id FROM channels WHERE created_by = ?))", userID).Error; err != nil {
			return err
		}
		if err := tx.Exec("DELETE FROM comments WHERE post_id IN (SELECT id FROM posts WHERE user_id = ?)", userID).Error; err != nil {
			return err
		}
		if err := tx.Exec("DELETE FROM posts WHERE user_id = ?", userID).Error; err != nil {
			return err
		}
		if err := tx.Exec("DELETE FROM comments WHERE user_id = ?", userID).Error; err != nil {
			return err
		}

		// Everything inside channels the user opened goes with it.
		if err := tx.Exec("DELETE FROM comments WHERE post_id IN (SELECT id FROM posts WHERE channel_id IN (SELECT id FROM channels WHERE created_by = ?))", userID).Error; err != nil {
			return err
		}
		if err := tx.Exec("DELETE FROM posts WHERE channel_id IN (SELECT id FROM channels WHERE created_by = ?)", userID).Error; err != nil {
			return err
		}
		if err := tx.Exec("DELETE FROM channels WHERE created_by = ?", userID).Error; err != nil {
			return err
		}

		return tx.Delete(&user).Error
	})
}

// EnsureAdmin creates the bootstrap admin account if it does not exist yet, or
// promotes the existing account to admin. The name is the account's identity
// now (there is no email), so a deployment that wants a pristine admin should
// point ADMIN_NAME at a name nobody else has registered: an existing
// non-admin account holding that name would be promoted.
func (s *UserService) EnsureAdmin(ctx context.Context, name, password string) (*models.User, error) {
	var user models.User

	err := s.db.WithContext(ctx).Where("lower(name) = lower(?)", name).First(&user).Error
	switch {
	case err == nil:
		if slices.Contains(user.Roles, "admin") {
			return &user, nil
		}
		user.Roles = append(user.Roles, "admin")
		if err := s.db.WithContext(ctx).Save(&user).Error; err != nil {
			return nil, err
		}
		return &user, nil
	case !errors.Is(err, gorm.ErrRecordNotFound):
		return nil, err
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	admin := models.User{
		Name:     name,
		Password: string(hashed),
		Roles:    []string{"admin"},
	}
	if err := s.db.WithContext(ctx).Create(&admin).Error; err != nil {
		return nil, err
	}
	return &admin, nil
}
