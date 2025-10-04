import { createBrowserRouter } from "react-router-dom";
import MainLayout from "~/layouts/main";
import Home from "~/pages/home";
import NotFound from "~/pages/not-found";
import Profile from "~/pages/profile";

const routes =  createBrowserRouter([
    {
        path: '/',
        element: <MainLayout />,
        children: [
            {
                index: true,
                element: <Home />
            },
            {
                path: "profile/:id",
                element: <Profile />,
            },
            {
                path: '*',
                element: <NotFound />
            }
        ]
    },
])

export default routes