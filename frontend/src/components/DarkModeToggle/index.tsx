import { MoonIcon, SunIcon } from "@/icons";
import styles from "./styles.module.css";
import "./icons.css";

export default function DarkModeToggle() {
  const setTheme = (theme: "dark" | "light") => {
    localStorage.setItem("theme", theme);
    document.documentElement.setAttribute("data-theme", theme);
  };

  const storedTheme = localStorage.getItem("theme");

  const prefersDark =
    window.matchMedia &&
    window.matchMedia("(prefers-color-scheme: dark)").matches;

  const defaultDark =
    storedTheme === "dark" || storedTheme === null && prefersDark;

  if (defaultDark) setTheme("dark");

  const toggleDarkMode = (e: Event) => {
    const target = e.target as HTMLInputElement;
    setTheme(target.checked ? "dark" : "light");
  };

  return (
    <div class={styles.overview__switch_wrapper}>
      <label class={styles.overview__theme_switch} for="checkbox">
        <input
          type="checkbox"
          id="checkbox"
          onChange={toggleDarkMode}
          checked={defaultDark}
        />
        <div class="overview__slider">
          <SunIcon class="sun-icon" />
          <MoonIcon class="moon-icon" />
        </div>
      </label>
    </div>
  );
}
