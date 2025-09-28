import { useRef } from 'react';
import sendSvg from '~/components/Input/send.svg';

export default function Input() {
  const textareaRef = useRef(null);

  const handleChange = () => {
    const textarea = textareaRef.current;
    if (textarea) {
      textarea.style.height = 'auto';
      textarea.style.height = textarea.scrollHeight + 'px';

      window.scrollTo({
        top: document.body.scrollHeight,
        behavior: 'smooth',
      });
    }
  };

  return (
    <div className="w-full p-2 mt-0 sticky bottom-0 left-0 right-0 flex justify-center">
      <div className="w-full relative">
        <textarea
          ref={textareaRef}
          onChange={handleChange}
          className="w-full mx-1 shadow-md px-3 pr-13 py-3 text-base resize-none overflow-hidden
          border border-gray-400 rounded-md focus:outline-none field-sizing-content bg-white focus:border-blue-500 max-h-[300px]"
          placeholder="Message @channel"
        />
        <button className="hover:bg-gray-200 p-2 right-1 top-1 mb-1 absolute rounded-xl">
          <img src={sendSvg} />
        </button>
      </div>
    </div>
  );
}
