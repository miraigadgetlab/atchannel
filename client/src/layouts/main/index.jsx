import { Outlet } from "react-router-dom";
import TopBar from '~/components/TopBar'

export default function MainLayout() {
    return (
        <main>
            <TopBar 
                title="@channel"
                boards={[
                    {
                        name: "/d",
                        href: 'https://discord.gg/E4ndGvEAqb'
                    },
                    {
                        name: "/g",
                        href: 'https://github.com/miraigadgetlab/atchannel'
                    },
                    {
                        name: "/auth",
                        href: 'https://atchannel.top/auth'
                    },
                    {
                        name: "/profil",
                        href: 'https://atchannel.top/auth'
                    }
                ]} />
            <Outlet />
        </main>
    )
}