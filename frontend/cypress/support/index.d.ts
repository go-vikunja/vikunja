/// <reference types="cypress" />

declare namespace Cypress {
  interface Chainable<Subject = any> {
    /**
     * Pastes a file onto the subject element.
     * @param fileName The name of the file to paste
     * @param fileType The MIME type of the file (defaults to 'image/png')
     */
    pasteFile(fileName: string, fileType?: string): Chainable<Subject>;
  }
} 
