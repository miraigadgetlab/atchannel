<script>
    export let avatar = "./avatar.png";
    export let id = "ID";
    export let nick = "Vaybe";
    export let sage = true;
    export let message = "text here";
    export let date = "Date";
    export let postNumber = 0;
</script>

<div class="flex space-x-4 m-2 items-start">
    <img src={avatar} alt="avatar" class="h-20 rounded-xl shadow-xl" />

    <div
        class="flex-1 p-2 rounded-xl shadow-xl relative speech-bubble
               {postNumber % 3 === 0
            ? 'bg-green-100 bubble-color-0'
            : postNumber % 3 === 1
              ? 'bg-blue-100 bubble-color-1'
              : 'bg-purple-200 bubble-color-2'}"
    >
        <div
            class="flex items-center space-x-3 text-xs sm:text-sm text-gray-800"
        >
            <span>{postNumber}</span>
            <a href="/p/{id}" class="text-black hover:text-blue-500 mx-4"
                >{nick}</a
            >
            {#if sage}
                <span class="text-gray-500">[sage]</span>
            {/if}
            <span class="ml-auto font-mono text-[10px] sm:text-xs">{id}</span>
        </div>

        <div class="text-black text-sm sm:text-base break-words break-all">
            {message}
        </div>

        <div class="text-[10px] sm:text-xs text-gray-400 mt-1 text-right">
            {date}
        </div>
    </div>
</div>

<style>
    .speech-bubble::before {
        content: "";
        position: absolute;
        left: -15px;
        top: 40px;
        border-top: 10px solid transparent;
        border-bottom: 10px solid transparent;
        border-right-width: 15px;
        transform: skewY(50deg);
    }

    .bubble-color-0::before {
        border-right-color: rgba(220, 252, 231, 1);
    }
    .bubble-color-1::before {
        border-right-color: rgba(219, 234, 254, 1);
    }
    .bubble-color-2::before {
        border-right-color: rgba(233, 213, 255, 1);
    }
</style>
