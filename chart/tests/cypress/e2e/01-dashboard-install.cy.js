before(function () {
    cy.visit(Cypress.env('grafana_url'));
  
    if (Cypress.env('keycloak_test_enable')) {
      cy.wait(500);
      cy.contains('SSO').click();
      cy.performKeycloakLogin(Cypress.env('tnr_username'), Cypress.env('tnr_password'))
    } else {
      cy.visit(Cypress.env('grafana_url'))
        .then(() => {
          // Check if the URL contains '/login'
          cy.url().then(currentUrl => {
            if (currentUrl.includes('/login')) {
              // Perform login if the URL contains '/login'
              cy.performGrafanaLogin(Cypress.env('grafana_username'), Cypress.env('grafana_password'));
            }
          });
        });
    }
  });
  
  describe('BBCTL Dashboard Testing', function () {
    it('Test for Logs Dashboard', function () {
      let dashboards = [
        "bbctl-all-logs-dashboard",
        "bbctl-policies-dashboard",
        "bbctl-preflight-dashboard",
        "bbctl-status-dashboard",
        "bbctl-version-dashboard",
        "bbctl-violations-dashboard"
      ]
      cy.visit(`${Cypress.env('grafana_url')}/dashboards`);
      cy.wait(1000);
      cy.get('input[placeholder="Search for dashboards and folders"]').type('-');
      cy.wait(500);
      for (const title of dashboards) {
        cy.get('a').contains(title);
      }
    });
  });
  