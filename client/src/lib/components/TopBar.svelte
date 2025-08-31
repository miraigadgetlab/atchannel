<script lang="ts">
    export let title: string = "@channel";

    export let boards: Array<{ name: string; href: string }> = [
        { name: "/discord", href: "https://discord.gg/E4ndGvEAqb" },
    ];
</script>

<header
    class="sticky top-0 flex items-center justify-between py-3 px-6 bg-white z-50"
>
    <div>
        <h1 class="text-xl font-bold text-black">{title}</h1>
    </div>

    <nav class="flex space-x-6 text-sm">
        {#each boards as board (board.name)}
            <a
                href={board.href}
                class="text-black font-medium hover:text-blue-400"
            >
                {board.name}
            </a>
        {/each}
    </nav>
</header>
