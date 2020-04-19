module.exports = {
  projectName: 'ORY Oathkeeper',
  projectSlug: 'oathkeeper',
  projectTagLine: 'A cloud native Identity & Access Proxy / API (IAP) and Access Control Decision API that authenticates, authorizes, and mutates incoming HTTP(s) requests. Inspired by the BeyondCorp / Zero Trust white paper. Written in Go.',
  updateTags: [
    {
      image: 'oryd/oathkeeper',
      files: [
        'docs/docs/install.md',
        'docs/docs/configure-deploy.md'
      ]
    },
    {
      replacer: ({content, next}) => content.replace(/v(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?/gi, `${next}`),
      files: [
        'docs/docs/install.md',
      ]
    }
  ],
  updateConfig: {
    src: '.schema/config.schema.json',
    dst: './docs/docs/reference/configuration.md'
  }
};
