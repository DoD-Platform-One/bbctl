
module.exports = {
    e2e: {
      env: {
        grafana_url: "https://grafana.dev.bigbang.mil",  
      },
      testIsolation: true,
      video: true,
      screenshot: true,
      supportFile: false,
      setupNodeEvents(on, config) {
        // implement node event listeners here
      },
    },
  }