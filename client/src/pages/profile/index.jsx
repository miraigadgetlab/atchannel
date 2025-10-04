import { useParams } from "react-router-dom";
import Post from "~/components/Post";

export default function Profile() {
  const { id } = useParams(); 
  id.a = "a";

  return (
    <div className="p-4 w-300">
      <div>
        <img className="rounded-tl-xl rounded-tr-xl sm:w-300 w-300 sm:h-96 h-48 bg-black" src="" />
        <img className="absolute top-45 sm:top-80 sm:ml-5 ml-2 w-32 h-32 sm:w-48 sm:h-48 border-white rounded-full border-8 flex bg-black" src="" />
        <h1 className="absolute sm:top-95 top-55 ml-38 sm:ml-55 text-white font-bold text-[24px] sm:text-[48px]">username</h1>
      </div>
      <div className="px-5 py-12">
      </div>
    </div>
  );
}
