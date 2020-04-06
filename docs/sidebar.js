module.exports = {
  Introduction:[
    "index",
    "install",
  ],
  "Core Concepts":[
    "api-access-rules",
    {
      type: "category",
      label: "Handlers",
      items:[
        "pipeline/index",
        "pipeline/authn",
        "pipeline/authz",
        "pipeline/mutator",
        "pipeline/error"
      ]
    },
  ],
  "Guides":["configure-deploy"],
  "Reference":[
    "reference/configuration",
    "reference/api"
  ],
  "SDKs":[
    "sdk/index"
  ],
};
