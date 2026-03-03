import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';

export default defineConfig({
  site: 'https://infrashift.github.io',
  base: '/zettelkasten',
  integrations: [
    starlight({
      title: 'Zettelkasten CLI',
      social: [
        {
          icon: 'github',
          label: 'GitHub',
          href: 'https://github.com/infrashift/zettelkasten',
        },
      ],
      editLink: {
        baseUrl: 'https://github.com/infrashift/zettelkasten/edit/main/docs/',
      },
      customCss: ['./src/styles/custom.css'],
      sidebar: [
        {
          label: 'Getting Started',
          items: [
            { label: 'Installation', slug: 'getting-started/installation' },
            { label: 'Quick Start', slug: 'getting-started/quick-start' },
            { label: 'Configuration', slug: 'getting-started/configuration' },
            { label: 'Note Format', slug: 'getting-started/note-format' },
          ],
        },
        {
          label: 'Methodology',
          items: [
            { label: 'Why Zettelkasten?', slug: 'methodology/why-zettelkasten' },
          ],
        },
        {
          label: 'Tutorials',
          items: [
            { label: 'NeoVim ZK Plugin Tutorial', slug: 'tutorial' },
            { label: 'ZK MCP Server Tutorial', slug: 'tutorial/mcp-server' },
          ],
        },
        {
          label: 'NeoVim Integration',
          items: [
            { label: 'Plugin Installation', slug: 'neovim/plugin-install' },
            { label: 'User Commands', slug: 'neovim/user-commands' },
            { label: 'Notes Workflow', slug: 'neovim/notes-workflow' },
            { label: 'Daily Notes Workflow', slug: 'neovim/daily-notes-workflow' },
            { label: 'Todo Workflow', slug: 'neovim/todo-workflow' },
          ],
        },
        {
          label: 'Integrations',
          items: [
            { label: 'MCP Server', slug: 'integrations/mcp-server' },
          ],
        },
        {
          label: 'Reference',
          items: [
            { label: 'CLI Commands', slug: 'reference/cli-commands' },
            { label: 'Project Structure', slug: 'reference/project-structure' },
          ],
        },
      ],
    }),
  ],
});
