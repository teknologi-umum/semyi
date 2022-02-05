import { Route, Routes } from "solid-app-router";
import OverviewPage from "@/pages/Overview";
import StatusPage from "@/pages/Status";

export default function App() {
  return (
    <Routes>
      <Route path="/" component={OverviewPage} />
      <Route path="/status" component={StatusPage} />
    </Routes>
  );
}
