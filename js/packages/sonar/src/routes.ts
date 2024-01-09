import { LoaderFunctionArgs, createBrowserRouter } from "react-router-dom";
import { Layout } from "./components/Layout.tsx";
import { Home } from "./components/Home.tsx";
import { Login } from "./components/Login.tsx";
import { redirect } from "react-router-dom";
import { Surveys } from "./components/Surveys.tsx";
import { Error } from "./components/Error.tsx";

function onlyAuthenticated({ request }: LoaderFunctionArgs) {
    if (!localStorage.getItem('token')) {
      let params = new URLSearchParams();
      params.set("from", new URL(request.url).pathname);
      return redirect("/login?" + params.toString());
    }
    return null;
  }

  function onlyUnauthenticated({ request }: LoaderFunctionArgs) {
    if (localStorage.getItem('token')) {
      return redirect("/");
    }
    return null;
  }

export const router = createBrowserRouter([
    {
      id: "root",
      path: "/",
      Component: Layout,
      children: [
        {
          index: true,
          Component: Home,
        },
        {
          path: "login",
          Component: Login,
          loader: onlyUnauthenticated,
        },
        {
          path: "surveys",
          loader: onlyAuthenticated,
          Component: Surveys,
        },
      ],
    },
    {
      path: "/logout",
      async action() {
        localStorage.removeItem('token');
        return redirect("/");
      },
    },
  ]);