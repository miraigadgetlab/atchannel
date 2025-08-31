<script lang="ts">
    import { onMount } from "svelte";

    let textarea: HTMLTextAreaElement;
    let message: string = "";

    export let onResize: (height: number) => void;

    function adjustHeight() {
        if (textarea) {
            textarea.style.height = "auto";
            const maxHeight = 300;
            if (textarea.scrollHeight > maxHeight) {
                textarea.style.height = maxHeight + "px";
                textarea.style.overflowY = "auto";
            } else {
                textarea.style.height = textarea.scrollHeight + "px";
                textarea.style.overflowY = "hidden";
            }

            onResize?.(textarea.offsetHeight);
        }
    }

    $: {
        message;
        adjustHeight();
    }

    onMount(() => {
        adjustHeight();
    });
</script>

<div class="fixed bottom-0 left-0 right-0 flex justify-center">
    <div class="bg-white w-300 px-4 py-2">
        <textarea
            bind:this={textarea}
            bind:value={message}
            on:input={adjustHeight}
            rows="1"
            placeholder="Message @channel"
            class="w-full px-3 py-2 text-base resize-none overflow-hidden
               border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 max-h-[300px]"
        ></textarea>
    </div>
</div>
