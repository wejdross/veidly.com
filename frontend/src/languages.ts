// Comprehensive list of all official EU languages plus Swiss languages
// Includes flag emoji for each language

export interface Language {
  code: string
  name: string
  flag: string
  nativeName?: string
}

export const LANGUAGES: Language[] = [
  // EU Official Languages (24 languages)
  { code: 'bg', name: 'Bulgarian', flag: '🇧🇬', nativeName: 'Български' },
  { code: 'hr', name: 'Croatian', flag: '🇭🇷', nativeName: 'Hrvatski' },
  { code: 'cs', name: 'Czech', flag: '🇨🇿', nativeName: 'Čeština' },
  { code: 'da', name: 'Danish', flag: '🇩🇰', nativeName: 'Dansk' },
  { code: 'nl', name: 'Dutch', flag: '🇳🇱', nativeName: 'Nederlands' },
  { code: 'en', name: 'English', flag: '🇬🇧', nativeName: 'English' },
  { code: 'et', name: 'Estonian', flag: '🇪🇪', nativeName: 'Eesti' },
  { code: 'fi', name: 'Finnish', flag: '🇫🇮', nativeName: 'Suomi' },
  { code: 'fr', name: 'French', flag: '🇫🇷', nativeName: 'Français' },
  { code: 'de', name: 'German', flag: '🇩🇪', nativeName: 'Deutsch' },
  { code: 'el', name: 'Greek', flag: '🇬🇷', nativeName: 'Ελληνικά' },
  { code: 'hu', name: 'Hungarian', flag: '🇭🇺', nativeName: 'Magyar' },
  { code: 'ga', name: 'Irish', flag: '🇮🇪', nativeName: 'Gaeilge' },
  { code: 'it', name: 'Italian', flag: '🇮🇹', nativeName: 'Italiano' },
  { code: 'lv', name: 'Latvian', flag: '🇱🇻', nativeName: 'Latviešu' },
  { code: 'lt', name: 'Lithuanian', flag: '🇱🇹', nativeName: 'Lietuvių' },
  { code: 'mt', name: 'Maltese', flag: '🇲🇹', nativeName: 'Malti' },
  { code: 'pl', name: 'Polish', flag: '🇵🇱', nativeName: 'Polski' },
  { code: 'pt', name: 'Portuguese', flag: '🇵🇹', nativeName: 'Português' },
  { code: 'ro', name: 'Romanian', flag: '🇷🇴', nativeName: 'Română' },
  { code: 'sk', name: 'Slovak', flag: '🇸🇰', nativeName: 'Slovenčina' },
  { code: 'sl', name: 'Slovenian', flag: '🇸🇮', nativeName: 'Slovenščina' },
  { code: 'es', name: 'Spanish', flag: '🇪🇸', nativeName: 'Español' },
  { code: 'sv', name: 'Swedish', flag: '🇸🇪', nativeName: 'Svenska' },

  // Swiss Languages (official)
  { code: 'rm', name: 'Romansh', flag: '🇨🇭', nativeName: 'Rumantsch' },

  // Commonly spoken in EU
  { code: 'tr', name: 'Turkish', flag: '🇹🇷', nativeName: 'Türkçe' },
  { code: 'ar', name: 'Arabic', flag: '🇸🇦', nativeName: 'العربية' },
  { code: 'ru', name: 'Russian', flag: '🇷🇺', nativeName: 'Русский' },
  { code: 'uk', name: 'Ukrainian', flag: '🇺🇦', nativeName: 'Українська' },
  { code: 'zh', name: 'Chinese', flag: '🇨🇳', nativeName: '中文' },
]

// Create a map for quick lookups
export const LANGUAGE_MAP: { [key: string]: Language } = LANGUAGES.reduce(
  (acc, lang) => ({
    ...acc,
    [lang.code]: lang,
  }),
  {}
)

// Get language display name with flag
export function getLanguageDisplay(code: string): string {
  const lang = LANGUAGE_MAP[code]
  if (!lang) return code
  return `${lang.flag} ${lang.name}`
}

// Get multiple language display names
export function getLanguagesDisplay(codes: string): string {
  return codes
    .split(',')
    .filter(code => code.trim())
    .map(code => getLanguageDisplay(code.trim()))
    .join(', ')
}

// For select options
export function getLanguageOption(code: string): string {
  const lang = LANGUAGE_MAP[code]
  if (!lang) return code
  return `${lang.flag} ${lang.name}`
}
