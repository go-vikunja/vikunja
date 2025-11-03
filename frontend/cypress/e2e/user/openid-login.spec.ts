context('OpenID Login', () => {
  it('logs in via Dex provider', () => {
    cy.visit('/login')
    cy.contains('Dex').click()
    cy.origin('http://dex:5556', () => {
      cy.get('#login').type('test@example.com')
      cy.get('#password').type('12345678')
      cy.get('#submit-login').click()
    })
    cy.url().should('include', '/')
    cy.get('main.app-content .content h2')
    	.should('contain', 'test!')
    cy.get('.show-tasks h3')
    	.should('contain', 'Current Tasks')
  })
})
