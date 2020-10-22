module.exports = {
  title: 'CassKop',
  tagline: 'Open-Source, Apache Cassandra operator for Kubernetes',
  url: 'https://orange-opensource.github.io',
  baseUrl: '/casskop/',
  onBrokenLinks: 'throw',
  favicon: 'img/casskop.ico',
  organizationName: 'Orange-OpenSource', // Usually your GitHub org/user name.
  projectName: 'casskop', // Usually your repo name.
  themeConfig: {
    navbar: {
      title: 'CassKop',
      logo: {
        alt: 'CassKop Logo',
        src: 'img/casskop_alone.png',
      },
      items: [
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
  presets: [
    [
      '@docusaurus/preset-classic',
      {
        docs: {
          sidebarPath: require.resolve('./sidebars.js'),
          // Please change this to your repo.
          editUrl:
            'https://github.com/Orange-OpenSource//casskop/edit/master/website/',
        },
        blog: {
          showReadingTime: true,
          // Please change this to your repo.
          editUrl:
            'https://github.com/Orange-OpenSource//casskop/edit/master/website/blog',
        },
        theme: {
          customCss: require.resolve('./src/css/custom.css'),
        },
      },
    ],
  ],
};
