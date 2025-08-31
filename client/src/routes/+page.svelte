<script>
    import Post from "$lib/components/Post.svelte";
    let posts = [
        {
            avatar: "https://i.pinimg.com/736x/90/dc/8f/90dc8f98920f6028b3015689415380f8.jpg",
            postNumber: 1,
            nick: "Kurisu",
            sage: true,
            id: "EN5001",
            message:
                "@channel why are we wasting time on jellied bananas again? serious question.",
            date: "2025/08/26 13:00:12",
        },
        {
            avatar: "https://i.pinimg.com/736x/ac/19/15/ac1915f262fdd06d179f685f3d0fb22f.jpg",
            postNumber: 2,
            nick: "Mayuri",
            sage: false,
            id: "EN5002",
            message:
                "bc they wobble and it's cute lol. dunno why u hate fun so much kurisu.",
            date: "2025/08/26 13:01:05",
        },
        {
            avatar: "https://cdn.discordapp.com/avatars/1257406444408672267/6272d21bb61adac4d71483eed03f79cc.webp?size=1024",
            postNumber: 3,
            nick: "Daru",
            sage: false,
            id: "EN5003",
            message:
                "cope harder, jellied bananas = literal proof of time distortion. normal food doesn’t just do that. checkmate.",
            date: "2025/08/26 13:02:30",
        },
        {
            avatar: "https://i.pinimg.com/736x/90/dc/8f/90dc8f98920f6028b3015689415380f8.jpg",
            postNumber: 4,
            nick: "Kurisu",
            sage: true,
            id: "EN5004",
            message:
                "bro wtf are you on. it's a side effect from the damn microwave experiments. stop making it sound like alien tech.",
            date: "2025/08/26 13:03:44",
        },
        {
            avatar: "https://i.pinimg.com/736x/ac/19/15/ac1915f262fdd06d179f685f3d0fb22f.jpg",
            postNumber: 5,
            nick: "Mayuri",
            sage: false,
            id: "EN5005",
            message:
                "okarin actually ate one btw. said texture was cursed af lol.",
            date: "2025/08/26 13:04:51",
        },
        {
            avatar: "https://cdn.discordapp.com/avatars/1257406444408672267/6272d21bb61adac4d71483eed03f79cc.webp?size=1024",
            postNumber: 6,
            nick: "Daru",
            sage: false,
            id: "EN5006",
            message:
                "lmao imagine risking ur stomach for SCIENCE. giga chad move tbh.",
            date: "2025/08/26 13:06:10",
        },
        {
            avatar: "https://i.pinimg.com/736x/5e/6b/ee/5e6beeb749ec884b9dc6ac4dd2657142.jpg",
            postNumber: 7,
            nick: "Okabe",
            sage: false,
            id: "EN5007",
            message:
                "FOOLS! you dare mock the sacred fruit of chaos? these bananas are not mere food items. they are the heralds of a distorted reality, proof that we, the chosen few, have bent the laws of causality itself. when the world trembles before the divergence, history will remember this day not as the hour of fear, but the moment mankind took its first step into the abyss. laugh all you want, but know this: the jellied banana is a prophecy made flesh. EL. PSY. KONGROO.",
            date: "2025/08/26 13:08:42",
        },
        {
            avatar: "https://i.pinimg.com/736x/90/dc/8f/90dc8f98920f6028b3015689415380f8.jpg",
            postNumber: 8,
            nick: "Kurisu",
            sage: true,
            id: "EN5008",
            message:
                "holy shit he’s actually roleplaying about bananas again. i can’t deal with this.",
            date: "2025/08/26 13:09:30",
        },
        {
            avatar: "https://cdn.discordapp.com/avatars/1257406444408672267/6272d21bb61adac4d71483eed03f79cc.webp?size=1024",
            postNumber: 9,
            nick: "Daru",
            sage: false,
            id: "EN5009",
            message: "based mad scientist rant. 10/10 would screencap.",
            date: "2025/08/26 13:10:05",
        },
        {
            avatar: "https://i.pinimg.com/736x/ac/19/15/ac1915f262fdd06d179f685f3d0fb22f.jpg",
            postNumber: 10,
            nick: "Mayuri",
            sage: false,
            id: "EN5010",
            message:
                "ehhh it’s kinda cool tho. maybe the bananas are trying to tell us smth?",
            date: "2025/08/26 13:11:12",
        },
    ];
</script>

<div
    class="flex-1 overflow-auto flex flex-col p-2 scroll-container scrollbar-thin"
>
    <div class="py-2">
        {#each posts as post}
            <Post {...post} />
        {/each}
    </div>
</div>
