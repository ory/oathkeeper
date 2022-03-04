export default {
  "title": "ORY Oathkeeper",
  "tagline": "A cloud native Identity & Access Proxy / API (IAP) and Access Control Decision API that authenticates, authorizes, and mutates incoming HTTP(s) requests. Inspired by the BeyondCorp / Zero Trust white paper. Written in Go.",
  "url": "https://www.ory.sh/",
  "baseUrl": "/oathkeeper/docs/",
  "favicon": "img/favico.png",
  "onBrokenLinks": "warn",
  "onBrokenMarkdownLinks": "warn",
  "organizationName": "ory",
  "projectName": "oathkeeper",
  "themeConfig": {
    "prism": {
      "theme": {
        "plain": {
          "color": "#393A34",
          "backgroundColor": "#f6f8fa"
        },
        "styles": [
          {
            "types": [
              "comment",
              "prolog",
              "doctype",
              "cdata"
            ],
            "style": {
              "color": "#999988",
              "fontStyle": "italic"
            }
          },
          {
            "types": [
              "namespace"
            ],
            "style": {
              "opacity": 0.7
            }
          },
          {
            "types": [
              "string",
              "attr-value"
            ],
            "style": {
              "color": "#e3116c"
            }
          },
          {
            "types": [
              "punctuation",
              "operator"
            ],
            "style": {
              "color": "#393A34"
            }
          },
          {
            "types": [
              "entity",
              "url",
              "symbol",
              "number",
              "boolean",
              "variable",
              "constant",
              "property",
              "regex",
              "inserted"
            ],
            "style": {
              "color": "#36acaa"
            }
          },
          {
            "types": [
              "atrule",
              "keyword",
              "attr-name",
              "selector"
            ],
            "style": {
              "color": "#00a4db"
            }
          },
          {
            "types": [
              "function",
              "deleted",
              "tag"
            ],
            "style": {
              "color": "#d73a49"
            }
          },
          {
            "types": [
              "function-variable"
            ],
            "style": {
              "color": "#6f42c1"
            }
          },
          {
            "types": [
              "tag",
              "selector",
              "keyword"
            ],
            "style": {
              "color": "#00009f"
            }
          }
        ]
      },
      "darkTheme": {
        "plain": {
          "color": "#F8F8F2",
          "backgroundColor": "#282A36"
        },
        "styles": [
          {
            "types": [
              "prolog",
              "constant",
              "builtin"
            ],
            "style": {
              "color": "rgb(189, 147, 249)"
            }
          },
          {
            "types": [
              "inserted",
              "function"
            ],
            "style": {
              "color": "rgb(80, 250, 123)"
            }
          },
          {
            "types": [
              "deleted"
            ],
            "style": {
              "color": "rgb(255, 85, 85)"
            }
          },
          {
            "types": [
              "changed"
            ],
            "style": {
              "color": "rgb(255, 184, 108)"
            }
          },
          {
            "types": [
              "punctuation",
              "symbol"
            ],
            "style": {
              "color": "rgb(248, 248, 242)"
            }
          },
          {
            "types": [
              "string",
              "char",
              "tag",
              "selector"
            ],
            "style": {
              "color": "rgb(255, 121, 198)"
            }
          },
          {
            "types": [
              "keyword",
              "variable"
            ],
            "style": {
              "color": "rgb(189, 147, 249)",
              "fontStyle": "italic"
            }
          },
          {
            "types": [
              "comment"
            ],
            "style": {
              "color": "rgb(98, 114, 164)"
            }
          },
          {
            "types": [
              "attr-name"
            ],
            "style": {
              "color": "rgb(241, 250, 140)"
            }
          }
        ]
      },
      "additionalLanguages": [
        "pug"
      ]
    },
    "announcementBar": {
      "id": "supportus",
      "content": "Sign up for <a href=\"https://ory.us10.list-manage.com/subscribe?u=ffb1a878e4ec6c0ed312a3480&id=f605a41b53&group[17097][16]=1\">important security announcements</a> and if you like ORY Oathkeeper give it a ⭐️ on <a target=\"_blank\" rel=\"noopener noreferrer\" href=\"https://github.com/ory/oathkeeper\">GitHub</a>!",
      "backgroundColor": "#fff",
      "textColor": "#000",
      "isCloseable": true
    },
    "algolia": {
      "apiKey": "8463c6ece843b377565726bb4ed325b0",
      "indexName": "ory",
      "contextualSearch": true,
      "searchParameters": {
        "facetFilters": [
          [
            "tags:oathkeeper",
            "tags:docs"
          ]
        ]
      },
      "appId": "BH4D9OD16A"
    },
    "navbar": {
      "hideOnScroll": true,
      "logo": {
        "alt": "ORY Oathkeeper",
        "src": "img/logo-oathkeeper.svg",
        "srcDark": "img/logo-oathkeeper.svg",
        "href": "https://www.ory.sh/oathkeeper"
      },
      "items": [
        {
          "to": "/",
          "activeBasePath": "/oathkeeper/docs/",
          "label": "Docs",
          "position": "left"
        },
        {
          "href": "https://www.ory.sh/docs",
          "label": "Ecosystem",
          "position": "left"
        },
        {
          "href": "https://www.ory.sh/blog",
          "label": "Blog",
          "position": "left"
        },
        {
          "href": "https://community.ory.sh",
          "label": "Forum",
          "position": "left"
        },
        {
          "href": "https://www.ory.sh/chat",
          "label": "Chat",
          "position": "left"
        },
        {
          "href": "https://github.com/ory/oathkeeper",
          "label": "GitHub",
          "position": "left"
        },
        {
          "type": "docsVersionDropdown",
          "position": "right",
          "dropdownActiveClassDisabled": true,
          "dropdownItemsAfter": [
            {
              "to": "/versions",
              "label": "All versions"
            }
          ],
          "dropdownItemsBefore": []
        }
      ]
    },
    "footer": {
      "style": "dark",
      "copyright": "Copyright © 2020 ORY GmbH",
      "links": [
        {
          "title": "Company",
          "items": [
            {
              "label": "Imprint",
              "href": "https://www.ory.sh/imprint"
            },
            {
              "label": "Privacy",
              "href": "https://www.ory.sh/privacy"
            },
            {
              "label": "Terms",
              "href": "https://www.ory.sh/tos"
            }
          ]
        }
      ]
    },
    "colorMode": {
      "defaultMode": "light",
      "disableSwitch": false,
      "respectPrefersColorScheme": false,
      "switchConfig": {
        "darkIcon": "🌜",
        "darkIconStyle": {},
        "lightIcon": "🌞",
        "lightIconStyle": {}
      }
    },
    "docs": {
      "versionPersistence": "localStorage"
    },
    "metadatas": [],
    "hideableSidebar": false
  },
  "plugins": [
    [
      "@docusaurus/plugin-content-docs",
      {
        "path": "docs",
        "sidebarPath": "/Users/foobar/workspace/go/src/github.com/ory/oathkeeper/docs/contrib/sidebar.js",
        "editUrl": "https://github.com/ory/oathkeeper/edit/master/docs",
        "routeBasePath": "/",
        "showLastUpdateAuthor": true,
        "showLastUpdateTime": true,
        "disableVersioning": false
      }
    ],
    "@docusaurus/plugin-content-pages",
    "/Users/foobar/workspace/go/src/github.com/ory/oathkeeper/docs/src/plugins/docusaurus-plugin-matamo/index.js",
    "@docusaurus/plugin-sitemap"
  ],
  "themes": [
    [
      "@docusaurus/theme-classic",
      {
        "customCss": "/Users/foobar/workspace/go/src/github.com/ory/oathkeeper/docs/src/css/theme.css"
      }
    ],
    "@docusaurus/theme-search-algolia"
  ],
  "baseUrlIssueBanner": true,
  "i18n": {
    "defaultLocale": "en",
    "locales": [
      "en"
    ]
  },
  "onDuplicateRoutes": "warn",
  "customFields": {},
  "presets": [],
  "titleDelimiter": "|",
  "noIndex": false
};