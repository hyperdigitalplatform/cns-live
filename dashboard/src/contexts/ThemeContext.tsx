import React, { createContext, useContext, useEffect, useState } from 'react';

type Theme = 'light' | 'dark';

interface ThemeContextType {
  theme: Theme;
  toggleTheme: () => void;
  setTheme: (theme: Theme) => void;
}

const ThemeContext = createContext<ThemeContextType | undefined>(undefined);

export function ThemeProvider({ children }: { children: React.ReactNode }) {
  const [theme, setThemeState] = useState<Theme>(() => {
    // Check localStorage first
    const savedTheme = localStorage.getItem('theme') as Theme | null;
    if (savedTheme) return savedTheme;

    // Check system preference
    if (window.matchMedia('(prefers-color-scheme: dark)').matches) {
      return 'dark';
    }

    return 'light';
  });

  useEffect(() => {
    const root = window.document.documentElement;
    root.classList.remove('light', 'dark');
    root.classList.add(theme);
    localStorage.setItem('theme', theme);
    console.log('[ThemeContext] Theme changed to:', theme);
    console.log('[ThemeContext] HTML classes:', root.className);

    // Debug: Check if dark mode selector matches
    const testElements = document.querySelectorAll('[class*="dark:bg-dark-base"]');
    console.log('[ThemeContext] Elements with dark:bg-dark-base class:', testElements.length);
    if (testElements.length > 0) {
      const firstEl = testElements[0] as HTMLElement;
      const computedBg = window.getComputedStyle(firstEl).backgroundColor;
      console.log('[ThemeContext] Computed background-color:', computedBg);
      console.log('[ThemeContext] Element classes:', firstEl.className);
    }
  }, [theme]);

  const toggleTheme = () => {
    console.log('[ThemeContext] Toggle clicked, current theme:', theme);
    setThemeState((prev) => {
      const newTheme = prev === 'light' ? 'dark' : 'light';
      console.log('[ThemeContext] Switching from', prev, 'to', newTheme);
      return newTheme;
    });
  };

  const setTheme = (newTheme: Theme) => {
    setThemeState(newTheme);
  };

  return (
    <ThemeContext.Provider value={{ theme, toggleTheme, setTheme }}>
      {children}
    </ThemeContext.Provider>
  );
}

export function useTheme() {
  const context = useContext(ThemeContext);
  if (context === undefined) {
    throw new Error('useTheme must be used within a ThemeProvider');
  }
  return context;
}
