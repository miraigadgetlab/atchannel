import { Outlet } from "react-router-dom";
import TopBar from '~/components/TopBar'

export default function MainLayout() {
    return (
        <main>
            <TopBar title="@channel" boards={[{name: "/discord", href: 'https://discord.gg/E4ndGvEAqb'}, {name: "/github", href: 'https://github.com/miraigadgetlab/atchannel'}]} />
            <Outlet />
        </main>
    )
}