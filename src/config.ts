import type { ThemeConfig } from './types'

export const themeConfig: ThemeConfig = {
  // SITE INFO ///////////////////////////////////////////////////////////////////////////////////////////
  site: {
    website: 'https://jameshartig.dev/', // Site domain
    title: 'James Hartig', // Site title
    author: 'James Hartig', // Author name
    description:
      'The technical blog of James Hartig. Deep dives into real-world engineering challenges and technology explorations.', // Site description
    language: 'en-US' // Default language
  },

  // GENERAL SETTINGS ////////////////////////////////////////////////////////////////////////////////////
  general: {
    contentWidth: '48rem', // Content area width
    centeredLayout: false, // Use centered layout (false for left-aligned)
    themeToggle: false, // Show theme toggle button (uses system theme by default)
    postListDottedDivider: false, // Show dotted divider in post list
    footer: true, // Show footer
    fadeAnimation: false // Enable fade animations
  },

  // DATE SETTINGS ///////////////////////////////////////////////////////////////////////////////////////
  date: {
    dateFormat: 'YYYY-MM-DD', // Date format: YYYY-MM-DD, MM-DD-YYYY, DD-MM-YYYY, MONTH DAY YYYY, DAY MONTH YYYY
    dateSeparator: '-', // Date separator: . - / (except for MONTH DAY YYYY and DAY MONTH YYYY)
    dateOnRight: true // Date position in post list (true for right, false for left)
  },

  // POST SETTINGS ///////////////////////////////////////////////////////////////////////////////////////
  post: {
    readingTime: false, // Show reading time in posts
    toc: false, // Show table of contents (when there is enough page width)
    imageViewer: false, // Enable image viewer
    copyCode: true, // Enable copy button in code blocks
    linkCard: false // Enable link card
  }
}
