import { MoonIcon, SunIcon } from "@/icons";
import styles from "./styles.module.css";

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
    storedTheme === "dark" || (storedTheme === null && prefersDark);

  if (defaultDark) setTheme("dark");

  const toggleDarkMode = (e: Event) => {
    setTheme((e.target as HTMLInputElement).checked ? "dark" : "light");
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
        <div class={styles.overview__slider}>
          <SunIcon className={styles["sun-icon"]} />
          <MoonIcon className={styles["moon-icon"]} />
        </div>
      </label>
    </div>
  );
}
