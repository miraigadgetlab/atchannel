import defaultAvatar1 from '~/components/Post/avatar1.png';
import defaultAvatar2 from '~/components/Post/avatar2.png';
import defaultAvatar3 from '~/components/Post/avatar3.png';
import '~/components/Post/index.css';

export default function PostBubble({
  name = 'USERNAME',
  ID = 'ID',
  avatar, 
  text = 'Message Text!!!',
  postNumber = 0,
  sage = true,
  date = 'DATE',
}) {
  const bubbleColor =
    postNumber % 3 === 0
      ? 'bg-green-100 bubble-color-0'
      : postNumber % 3 === 1
      ? 'bg-blue-100 bubble-color-1'
      : 'bg-purple-200 bubble-color-2';

  const avatarByColor =
    postNumber % 3 === 0
      ? defaultAvatar1
      : postNumber % 3 === 1
      ? defaultAvatar2
      : defaultAvatar3;
  const finalAvatar = avatar || avatarByColor;

  return (
    <div className="flex space-x-2 m-3 pl-2 items-start sm:space-x-4">
      <img
        src={finalAvatar}
        alt="avatar"
        className="h-18 w-18 sm:h-22 sm:w-22 rounded-lg sm:rounded-xl"
      />
      <div
        className={`flex-1 p-2 px-3 rounded-[14px] relative speech-bubble ${bubbleColor}`}
      >
        <div className="flex items-center space-x-5 text-xs sm:text-sm">
          <span className="text-gray-500">{postNumber}</span>
          <a href={`/profile/${ID}`} className="hover:text-blue-500 font-medium">
            {name}
          </a>
          {sage && (
            <span className="text-black text-xs sm:text-sm">[sage]</span>
          )}
          <span className="ml-auto font-mono text-[9px] sm:text-xs text-gray-500">
            {ID}
          </span>
        </div>
        <div className="text-sm sm:text-base break-words break-all mt-1">
          {text}
        </div>
        <div className="text-[9px] sm:text-xs mt-1 text-right text-gray-600">
          {date}
        </div>
      </div>
    </div>
  );
}
