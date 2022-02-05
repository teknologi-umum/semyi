import { createApp } from "vue";
import { createRouter, createWebHistory } from "vue-router";
import Layout from "@/components/Layout.vue";
import OverviewPage from "@/pages/Overview.vue";
import StatusPage from "@/pages/Status.vue";
import "@/global.css";

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: "/", component: OverviewPage },
    { path: "/status", component: StatusPage },
  ],
});

createApp(Layout).use(router).mount("#app");
