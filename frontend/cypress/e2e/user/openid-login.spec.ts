context('OpenID Login', () => {
  it('logs in via Dex provider', () => {
    cy.visit('/login')
    cy.contains('Dex').click()
    cy.origin('http://dex:5556', () => {
      cy.get('#login').type('test')
      cy.get('#password').type('12345678')
      cy.get('#submit-login').click()
    })
    cy.url().should('include', '/')
    cy.get('h2').should('contain', 'Hi test!')
  })
})
