import { render } from "solid-js/web";
import "@/index.css";

import "@fontsource/ibm-plex-sans/500.css";
import "@fontsource/ibm-plex-sans/700.css";
import "@fontsource/libre-franklin/400.css";
import DetailPage from "@/pages/Detail";
import OverviewPage from "@/pages/Overview";
import { Route, Router } from "@solidjs/router";

render(
  () => (
    <Router>
      <Route path="/" component={OverviewPage} />
      <Route path="/by" component={DetailPage} />
    </Router>
  ),
  document.getElementById("root") as HTMLElement,
);
