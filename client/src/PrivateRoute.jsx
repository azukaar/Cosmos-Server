import { Suspense } from "react"
import { Await, Navigate } from "react-router"
import isLoggedIn from "./isLoggedIn"

function PrivateRoute({ children }) {
  return <Suspense>
    <Await resolve={isLoggedIn()}>
      {authStatus => {
        switch (authStatus) {
          case "OK": return children
          default: return <Navigate to={authStatus} replace />
        }
      }}
    </Await>
  </Suspense>
}

export default PrivateRoute
