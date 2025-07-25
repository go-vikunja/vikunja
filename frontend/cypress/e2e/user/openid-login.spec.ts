context('OpenID Login', () => {
  it('logs in via Dex provider', () => {
    cy.visit('/login')
    cy.contains('Dex').click()
    cy.origin('http://dex:5556', () => {
      cy.get('input[name="login"]').type('test')
      cy.get('input[name="password"]').type('password')
      cy.get('button[type="submit"]').click()
    })
    cy.url().should('include', '/')
    cy.get('h2').should('contain', 'Hi test!')
  })
})
