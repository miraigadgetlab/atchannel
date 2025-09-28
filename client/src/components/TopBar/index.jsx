export default function TopBar({ title, boards }) {
  return (
    <header className="bg-white p-2 sticky top-0 flex items-center justify-between py-3 z-50">
      <div>
        <h1 className="text-xl font-bold">{title}</h1>
      </div>
      <nav className="flex space-x-6 text-sm">
        {boards.map((board) => (
          <a
            key={board.name}
            href={board.href}
            className="font-medium hover:text-blue-400"
          >
            {board.name}
          </a>
        ))}
      </nav>
    </header>
  );
}
