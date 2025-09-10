## Issue Summary: Link Share Issues ([#1380](https://github.com/go-vikunja/vikunja/issues/1380))

### Description

Users experience inconsistent behavior when accessing shared links in Vikunja, specifically when entering the password. Behaviors observed include:

- Sometimes greeted with "Hallo Link Share" and only the task list, without the project name or ability to switch views.
- On entering the correct password, sometimes denied access (403 error) with no visual feedback; the screen goes blank. Wrong passwords also result in a 403 but with no feedback.
- Occasionally works as expected, showing the project name and allowing view switching.

Additional UX feedback:
- After selecting a task, returning to the board is unintuitive. The project name at the top should be clickable, and the project name below a task should be highlighted or have an icon for easier navigation.

Related to: [#524](https://github.com/go-vikunja/vikunja/issues/524)

### Version
- Occurs on latest/try version
- Reproducible on the Vikunja demo site

### Screenshots

![Link Share without project name](https://github.com/user-attachments/assets/2d9b7d24-4bbf-48bd-8993-62843adb1c6e)
![403 error, no feedback](https://github.com/user-attachments/assets/a193f004-ea05-4c9f-b7ac-50605aa5236d)
![Task view navigation issue](https://github.com/user-attachments/assets/c82b62bb-a4e5-4ff1-846c-6619ecb301d6)
