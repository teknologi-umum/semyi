import styles from "./DarkMode.module.css";
import cx from "classnames";

export default function DarkMode() {
  const svgFeatherMoon = cx(`${styles.feather__moon}`, "feather");
  const svgFeatherSun = cx(`${styles.feather__sun}`, "feather");
  const setDark = () => {
    localStorage.setItem("theme", "dark");

    document.documentElement.setAttribute("data-theme", "dark");
  };

  const setLight = () => {
    localStorage.setItem("theme", "light");
    document.documentElement.setAttribute("data-theme", "light");
  };

  const storedTheme = localStorage.getItem("theme");

  const prefersDark =
    window.matchMedia &&
    window.matchMedia("(prefers-color-scheme: dark)").matches;

  const defaultDark =
    storedTheme === "dark" || storedTheme === null && prefersDark;

  if (defaultDark) {
    setDark();
  }

  const toggleTheme = (e: any) => {
    if (e.target.checked) {
      setDark();
    } else {
      setLight();
    }
  };

  // defaultChecked={defaultDark}
  return (
    <div class={styles.overview__switch_wrapper}>
      <label class={styles.overview__theme_switch} for="checkbox">
        <input
          type="checkbox"
          id="checkbox"
          onChange={toggleTheme}
          checked={defaultDark}
        />
        <div class={styles.overview__slider}>
          <svg
            xmlns="http://www.w3.org/2000/svg"
            width="16"
            height="16"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
            class={svgFeatherSun}
          >
            <circle cx="12" cy="12" r="5"></circle>
            <line x1="12" y1="1" x2="12" y2="3"></line>
            <line x1="12" y1="21" x2="12" y2="23"></line>
            <line x1="4.22" y1="4.22" x2="5.64" y2="5.64"></line>
            <line x1="18.36" y1="18.36" x2="19.78" y2="19.78"></line>
            <line x1="1" y1="12" x2="3" y2="12"></line>
            <line x1="21" y1="12" x2="23" y2="12"></line>
            <line x1="4.22" y1="19.78" x2="5.64" y2="18.36"></line>
            <line x1="18.36" y1="5.64" x2="19.78" y2="4.22"></line>
          </svg>
          <svg
            xmlns="http://www.w3.org/2000/svg"
            width="16"
            height="16"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
            class={svgFeatherMoon}
          >
            <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"></path>
          </svg>
        </div>
      </label>
    </div>
  );
}
