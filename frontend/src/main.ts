import "./app.css";
import { mount } from "svelte";
import App from "./App.svelte";
import { registerOfflineSupport } from "./offline";

registerOfflineSupport();

const app = mount(App, {
  target: document.getElementById("app")!
});

export default app;
