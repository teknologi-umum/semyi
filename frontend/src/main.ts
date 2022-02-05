import { createApp } from "vue";
import { createRouter, createWebHistory } from "vue-router";
import Layout from "@/components/Layout.vue";
import IndexPage from "@/pages/index.vue";
import StatusPage from "@/pages/status.vue";

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: "/", component: IndexPage },
    { path: "/status", component: StatusPage },
  ],
});

createApp(Layout).use(router).mount("#app");
