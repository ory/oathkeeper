module.exports = {
    projectName: 'ORY GTM Docs',
    projectSlug: 'gtmdoc',
    projectTagLine:
      'Ory GTM documentation.',
    updateTags: [
      {
        image: 'oryd/keto',
        files: ['docs/docs/configure-deploy.md']
      }
    ],
    updateConfig: {
      src: '.schema/config.schema.json',
      dst: './docs/docs/reference/configuration.md'
    }
  }