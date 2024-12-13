before(function () {
    cy.visit(Cypress.env('grafana_url'));
  
    if (Cypress.env('keycloak_test_enable')) {
      cy.wait(500);
      cy.contains('SSO').click();
      cy.performKeycloakLogin(Cypress.env('tnr_username'), Cypress.env('tnr_password'))
    } else {
      cy.get('input[name="user"]').type(Cypress.env('grafana_username'));
      cy.get('input[name="password"]').type(Cypress.env('grafana_password'));
      cy.contains('Log in').click();
      cy.wait(1000);
    }
  });
  
  describe('BBCTL Dashboard Testing', function () {
    it('Test for Logs Dashboard', function () {
      let dashboards = [
        "all-logs-dashboard",
        "policies-dashboard",
        "preflight-dashboard",
        "status-dashboard",
        "version-dashboard",
        "violations-dashboard"
      ]
      cy.visit(`${Cypress.env('grafana_url')}/dashboards`);
      cy.wait(1000);
      cy.get('div[data-testid="input-wrapper"]').within(() => {
        cy.get('input').type('-');
      })
      cy.wait(500);
      for (const title of dashboards) {
        cy.get('a').contains(title);
      }
    });
  });
  