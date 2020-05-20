module.exports = {
  title: 'Casskop',
  tagline: 'Open-Source, Apache Cassandra operator for Kubernetes',
  url: 'https://orange-opensource.github.io',
  baseUrl: '/casskop/',
  favicon: 'img/casskop.ico',
  organizationName: 'Orange-OpenSource', // Usually your GitHub org/user name.
  projectName: 'casskop', // Usually your repo name.
  themeConfig: {
    navbar: {
      title: 'Casskop',
      logo: {
        alt: 'Casskop Logo',
        src: 'img/casskop_alone.png',
      },
      links: [
        {to: 'docs/1_concepts/1_introduction', label: 'Docs', position: 'right'},
        {to: 'blog', label: 'Blog', position: 'right'},
        {
          href: 'https://github.com/Orange-OpenSource/casskop',
          label: 'GitHub',
          position: 'right',
        },
      ],
    },
    footer: {
      style: 'dark',
      links: [
        {
          title: 'Getting Started',
          items: [
            {
              label: 'Documentation',
              to: 'docs/1_concepts/1_introduction',
            },
            {
              label: 'GitHub',
              href: 'https://github.com/Orange-OpenSource/casskop',
            },
          ],
        },
        {
          title: 'Community',
          items: [
            {
              label: 'Slack',
              href: 'https://casskop.slack.com',
            },
            {
              label: 'Blog',
              to: 'blog',
            },
            {
              label: 'Twitter',
              href: 'https://twitter.com',
            },
          ],
        },
        {
          title: 'Contact',
          items: [
            {
              label: 'Email',
              href: 'mailto:prj.casskop.support@list.orangeportails.net',
            },
            {
              label: 'Feature request',
              href: 'https://github.com/Orange-OpenSource/casskop/issues',
            },
          ],
        },
      ],
      copyright: `Copyright © ${new Date().getFullYear()} Orange, Inc. Built with Docusaurus.`,
    },
  },
  themes: ['@docusaurus/theme-classic', '@docusaurus/theme-live-codeblock'],
  presets: [
    [
      '@docusaurus/preset-classic',
      {
        docs: {
          sidebarPath: require.resolve('./sidebars.js'),
          editUrl:
              'https://github.com/Orange-OpenSource//casskop/edit/master/website/',
        },
        theme: {
          customCss: require.resolve('./src/css/custom.css'),
        },
      },
    ],
  ],
};


