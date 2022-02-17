import { Route, Routes } from "solid-app-router";
import OverviewPage from "@/pages/Overview";
import DetailPage from "@/pages/Detail";

export default function App() {
  return (
    <Routes>
      <Route path="/" component={OverviewPage} />
      <Route path="/by" component={DetailPage} />
    </Routes>
  );
}
