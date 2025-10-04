import { useRef } from 'react';
import sendSvg from '~/components/Input/send.svg';
import emoteSvg from '~/components/Input/smile-plus.svg';
import gifSvg from '~/components/Input/gif.svg';

export default function Input() {
  const textareaRef = useRef(null);

  const handleChange = () => {
    const textarea = textareaRef.current;
    if (textarea) {
      textarea.style.height = 'auto';
      textarea.style.height = textarea.scrollHeight + 'px';

      window.scrollTo({
        top: document.body.scrollHeight,
        behavior: 'auto',
      });
    }
  };

  const handleSubmit = () => {
    const message = textareaRef.current.value.trim();
    if (message) {
      console.log('Mesaj gönderildi:', message);
      
      textareaRef.current.value = '';
      textareaRef.current.style.height = 'auto';
    }
  };

  const handleKeyDown = (e) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSubmit();
    }
  };

  return (
    <div className="w-full p-2 mt-0 sticky bottom-0 left-0 right-0 flex justify-center">
      <div className="w-full relative">
        <textarea
          ref={textareaRef}
          rows={1}
          onChange={handleChange}
          onKeyDown={handleKeyDown} 
          className="w-full mx-1 shadow-md px-3 pr-13 py-3 text-base resize-none overflow-hidden
          border border-[#2f3035] rounded-md focus:outline-none field-sizing-content bg-white focus:border-blue-500 min-h-[50px] max-h-[300px]"
          placeholder="Message @channel"
        />

        
        <div className='absolute right-1 top-1 flex'>
          <button
            onClick={handleSubmit} 
            className="
              m-1 p-1 rounded-md
              hover:bg-[#e6e6e8]
              transition-colors duration-250 ease-in-out
            "
          >
            <img src={gifSvg} alt="Emote" />
          </button>
          <button
            onClick={handleSubmit} 
            className="
              m-1 p-1 rounded-md
              hover:bg-[#e6e6e8]
              transition-colors duration-250 ease-in-out
            "
          >
            <img src={emoteSvg} alt="Emote" />
          </button>
          <button
            onClick={handleSubmit} 
            className="
              my-1 p-1 rounded-md
              hover:bg-[#e6e6e8]
              transition-colors duration-250 ease-in-out
            "
          >
            <img src={sendSvg} alt="Send" />
          </button>
        </div>
      </div>
    </div>
  );
}
