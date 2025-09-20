
> vikunja-frontend@0.10.0 lint /home/claude-testing/vikunja/frontend
> eslint 'src/**/*.{js,ts,vue}'


/home/claude-testing/vikunja/frontend/src/components/base/BaseButton.story.vue
  7:35  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any

/home/claude-testing/vikunja/frontend/src/components/home/ProjectsNavigation.vue
  90:97   error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  90:151  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any

/home/claude-testing/vikunja/frontend/src/components/input/AutocompleteDropdown.vue
  122:63  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any

/home/claude-testing/vikunja/frontend/src/components/input/Button.vue
  70:45  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any

/home/claude-testing/vikunja/frontend/src/components/input/DatepickerInline.vue
  153:40  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  160:40  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any

/home/claude-testing/vikunja/frontend/src/components/input/Multiselect.vue
  140:25  error  Unexpected any. Specify a different type                                                                                   @typescript-eslint/no-explicit-any
  142:1   error  defineProps should be the first statement in `<script setup>` (after any potential import statements or type definitions)  vue/define-macros-order
  383:63  error  Unexpected any. Specify a different type                                                                                   @typescript-eslint/no-explicit-any
  403:48  error  Unexpected any. Specify a different type                                                                                   @typescript-eslint/no-explicit-any
  471:41  error  Unexpected any. Specify a different type                                                                                   @typescript-eslint/no-explicit-any

/home/claude-testing/vikunja/frontend/src/components/input/Reactions.vue
  100:44  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  114:51  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any

/home/claude-testing/vikunja/frontend/src/components/input/editor/CommandsList.vue
  90:11  error  Missing trailing comma  comma-dangle

/home/claude-testing/vikunja/frontend/src/components/input/editor/commands.ts
  13:48  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  13:60  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  13:72  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any

/home/claude-testing/vikunja/frontend/src/components/input/editor/setLinkInEditor.ts
  3:44  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  3:57  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any

/home/claude-testing/vikunja/frontend/src/components/input/editor/suggestion.ts
   14:42  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
   14:54  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
   27:42  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
   27:54  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
   40:42  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
   40:54  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
   53:42  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
   53:54  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
   66:42  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
   66:54  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
   79:42  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
   79:54  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
   92:42  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
   92:54  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  105:42  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  105:54  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  118:42  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  118:54  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  131:42  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  131:54  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  144:42  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  144:54  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  175:22  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any

/home/claude-testing/vikunja/frontend/src/components/input/filter/FilterAutocomplete.ts
  188:81   error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  348:72   error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  348:102  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  380:61   error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any

/home/claude-testing/vikunja/frontend/src/components/input/filter/highlighter.ts
  248:31  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  253:25  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any

/home/claude-testing/vikunja/frontend/src/components/misc/CreateEdit.vue
  55:11  error  'CreateEditProps' is defined but never used. Allowed unused vars must match /^_/u                                          @typescript-eslint/no-unused-vars
  66:7   error  'props' is assigned a value but never used. Allowed unused vars must match /^_/u                                           @typescript-eslint/no-unused-vars
  75:9   error  Unexpected any. Specify a different type                                                                                   @typescript-eslint/no-explicit-any
  77:1   error  defineEmits should be the first statement in `<script setup>` (after any potential import statements or type definitions)  vue/define-macros-order

/home/claude-testing/vikunja/frontend/src/components/misc/Dropdown.vue
  51:7   error  'props' is assigned a value but never used. Allowed unused vars must match /^_/u                                           @typescript-eslint/no-unused-vars
  51:47  error  Unexpected any. Specify a different type                                                                                   @typescript-eslint/no-explicit-any
  53:1   error  defineEmits should be the first statement in `<script setup>` (after any potential import statements or type definitions)  vue/define-macros-order

/home/claude-testing/vikunja/frontend/src/components/notifications/Notifications.vue
  156:16  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  159:32  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any

/home/claude-testing/vikunja/frontend/src/components/project/ProjectWrapper.vue
  83:83  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any

/home/claude-testing/vikunja/frontend/src/components/project/views/ProjectKanban.vue
  428:59  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  491:38  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any

/home/claude-testing/vikunja/frontend/src/components/project/views/ProjectList.vue
  159:2   error  Getting a value from the ref object in the same scope will cause the value to lose reactivity  vue/no-ref-object-reactivity-loss
  159:35  error  Getting a value from the ref object in the same scope will cause the value to lose reactivity  vue/no-ref-object-reactivity-loss
  159:91  error  Missing trailing comma                                                                         comma-dangle
  239:36  error  Unexpected any. Specify a different type                                                       @typescript-eslint/no-explicit-any

/home/claude-testing/vikunja/frontend/src/components/project/views/ProjectTable.vue
  380:43  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  381:17  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  382:32  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any

/home/claude-testing/vikunja/frontend/src/components/project/views/ViewEditForm.vue
  29:14  error  Getting a value from the `props` in root scope of `<script setup>` will cause the value to lose reactivity  vue/no-setup-props-reactivity-loss

/home/claude-testing/vikunja/frontend/src/components/quick-actions/QuickActions.vue
  129:14  error  'IAbstract' is defined but never used. Allowed unused vars must match /^_/u  @typescript-eslint/no-unused-vars
  228:9   error  Unexpected any. Specify a different type                                     @typescript-eslint/no-explicit-any
  416:10  error  Unexpected any. Specify a different type                                     @typescript-eslint/no-explicit-any
  446:13  error  Unexpected any. Specify a different type                                     @typescript-eslint/no-explicit-any
  459:50  error  Unexpected any. Specify a different type                                     @typescript-eslint/no-explicit-any

/home/claude-testing/vikunja/frontend/src/components/sharing/UserTeam.vue
  209:50  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  212:47  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  263:51  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  279:60  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  280:22  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  280:48  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  292:42  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  325:42  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  331:34  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  339:17  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  347:52  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  356:27  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  374:46  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  376:46  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  385:43  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any

/home/claude-testing/vikunja/frontend/src/components/tasks/AddTask.vue
  167:1   error  Expected indentation of 5 tabs but found 4  indent
  168:1   error  Expected indentation of 4 tabs but found 3  indent
  258:14  error  Unexpected any. Specify a different type    @typescript-eslint/no-explicit-any

/home/claude-testing/vikunja/frontend/src/components/tasks/partials/Attachments.vue
  272:1  error  Expected indentation of 2 tabs but found 1  indent

/home/claude-testing/vikunja/frontend/src/components/tasks/partials/Comments.vue
  340:72   error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  340:138  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any

/home/claude-testing/vikunja/frontend/src/components/tasks/partials/Description.vue
  137:17  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any

/home/claude-testing/vikunja/frontend/src/components/tasks/partials/EditAssignees.vue
  47:14  error  'IAbstract' is defined but never used. Allowed unused vars must match /^_/u  @typescript-eslint/no-unused-vars

/home/claude-testing/vikunja/frontend/src/components/tasks/partials/KanbanCard.vue
  162:52  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any

/home/claude-testing/vikunja/frontend/src/components/tasks/partials/Reminders.story.vue
  17:15  warning  Attribute ':modelValue' must be hyphenated  vue/attribute-hyphenation

/home/claude-testing/vikunja/frontend/src/composables/useRouteWithModal.ts
  21:36  error  Missing trailing comma                    comma-dangle
  58:50  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any

/home/claude-testing/vikunja/frontend/src/composables/useTaskList.ts
   10:14  error  'IBucket' is defined but never used. Allowed unused vars must match /^_/u  @typescript-eslint/no-unused-vars
   54:90  error  Missing trailing comma                                                     comma-dangle
  133:63  error  Missing trailing comma                                                     comma-dangle

/home/claude-testing/vikunja/frontend/src/histoire.setup.ts
  25:38  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any

/home/claude-testing/vikunja/frontend/src/i18n/useDayjsLanguageSync.ts
  64:55  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any

/home/claude-testing/vikunja/frontend/src/main.ts
  65:37  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any

/home/claude-testing/vikunja/frontend/src/message/index.ts
   4:33  error  Unexpected any. Specify a different type                                                                             @typescript-eslint/no-explicit-any
  37:26  error  Unexpected any. Specify a different type                                                                             @typescript-eslint/no-explicit-any
  40:3   error  Use "@ts-expect-error" instead of "@ts-ignore", as "@ts-ignore" will do nothing if the following line is error-free  @typescript-eslint/ban-ts-comment
  47:28  error  Unexpected any. Specify a different type                                                                             @typescript-eslint/no-explicit-any
  50:3   error  Use "@ts-expect-error" instead of "@ts-ignore", as "@ts-ignore" will do nothing if the following line is error-free  @typescript-eslint/ban-ts-comment

/home/claude-testing/vikunja/frontend/src/models/abstractModel.ts
   6:45  error  'Model' is defined but never used. Allowed unused vars must match /^_/u  @typescript-eslint/no-unused-vars
  17:34  error  Unexpected any. Specify a different type                                 @typescript-eslint/no-explicit-any

/home/claude-testing/vikunja/frontend/src/models/backgroundImage.ts
  12:1   error  Expected indentation of 3 tabs but found 2  indent
  13:1   error  Expected indentation of 3 tabs but found 2  indent
  13:17  error  Missing trailing comma                      comma-dangle
  14:1   error  Expected indentation of 2 tabs but found 1  indent

/home/claude-testing/vikunja/frontend/src/models/notification.ts
   93:48   error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  102:56   error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  113:62   error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  118:51   error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  139:56   error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  144:102  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any

/home/claude-testing/vikunja/frontend/src/models/task.ts
  110:48  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  121:50  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  122:54  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  123:50  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  168:24  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  169:40  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  169:73  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any

/home/claude-testing/vikunja/frontend/src/models/taskComment.ts
  29:24  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  30:40  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  30:74  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any

/home/claude-testing/vikunja/frontend/src/models/taskReminder.ts
  14:52  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  15:53  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any

/home/claude-testing/vikunja/frontend/src/sentry.ts
  21:56  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any

/home/claude-testing/vikunja/frontend/src/services/abstractService.ts
    5:8   error  'AbstractModel' is defined but never used. Allowed unused vars must match /^_/u  @typescript-eslint/no-unused-vars
  357:14  error  Unexpected any. Specify a different type                                         @typescript-eslint/no-explicit-any
  485:56  error  Unexpected any. Specify a different type                                         @typescript-eslint/no-explicit-any

/home/claude-testing/vikunja/frontend/src/services/attachment.ts
  39:27  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  41:70  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any

/home/claude-testing/vikunja/frontend/src/services/backgroundUnsplash.ts
   3:8   error  'ProjectModel' is defined but never used. Allowed unused vars must match /^_/u  @typescript-eslint/no-unused-vars
  18:27  error  Unexpected any. Specify a different type                                        @typescript-eslint/no-explicit-any
  22:21  error  Unexpected any. Specify a different type                                        @typescript-eslint/no-explicit-any

/home/claude-testing/vikunja/frontend/src/services/backgroundUpload.ts
  5:15  error  'IFile' is defined but never used. Allowed unused vars must match /^_/u  @typescript-eslint/no-unused-vars

/home/claude-testing/vikunja/frontend/src/services/bucket.ts
  20:22  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  22:38  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any

/home/claude-testing/vikunja/frontend/src/services/label.ts
  17:22  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  28:22  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  32:22  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any

/home/claude-testing/vikunja/frontend/src/services/notification.ts
  24:69  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  26:68  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any

/home/claude-testing/vikunja/frontend/src/services/passwordReset.ts
  12:8   error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  13:19  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  23:57  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  33:57  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any

/home/claude-testing/vikunja/frontend/src/services/project.ts
  25:41  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any

/home/claude-testing/vikunja/frontend/src/services/reactions.ts
  20:57  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  21:17  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  25:48  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any

/home/claude-testing/vikunja/frontend/src/services/savedFilter.ts
  40:40  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any

/home/claude-testing/vikunja/frontend/src/services/task.ts
   11:26  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
   46:29  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
   71:32  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
   97:80  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  105:30  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  113:40  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  121:79  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any

/home/claude-testing/vikunja/frontend/src/services/taskCollection.ts
  48:21  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any

/home/claude-testing/vikunja/frontend/src/services/taskComment.ts
  31:75  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any

/home/claude-testing/vikunja/frontend/src/services/teamProject.ts
  4:8  error  'TeamModel' is defined but never used. Allowed unused vars must match /^_/u  @typescript-eslint/no-unused-vars

/home/claude-testing/vikunja/frontend/src/services/totp.ts
  22:16  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  26:17  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any

/home/claude-testing/vikunja/frontend/src/services/userProject.ts
  4:8  error  'UserModel' is defined but never used. Allowed unused vars must match /^_/u  @typescript-eslint/no-unused-vars

/home/claude-testing/vikunja/frontend/src/stores/auth.ts
  172:36  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  186:15  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  205:39  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  219:15  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  234:57  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  234:68  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  264:56  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  264:71  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  348:15  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  355:21  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  377:15  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  421:15  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  443:16  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any

/home/claude-testing/vikunja/frontend/src/stores/kanban.ts
  292:102  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  300:77   error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  306:57   error  Missing trailing comma                    comma-dangle
  347:67   error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  350:63   error  Missing trailing comma                    comma-dangle
  377:75   error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any

/home/claude-testing/vikunja/frontend/src/stores/tasks.ts
   42:38  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
   42:63  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
   43:40  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
   95:58  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  141:17  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  142:31  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any

/home/claude-testing/vikunja/frontend/src/sw.ts
   6:53  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  11:22  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  13:26  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  13:41  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  16:35  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  17:26  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  20:32  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  20:47  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  30:39  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  89:34  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any

/home/claude-testing/vikunja/frontend/src/views/Home.vue
  83:99  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any

/home/claude-testing/vikunja/frontend/src/views/migrate/MigrationHandler.vue
  141:11  error  'MigrationAuthResponse' is defined but never used. Allowed unused vars must match /^_/u  @typescript-eslint/no-unused-vars
  187:70  error  Unexpected any. Specify a different type                                                 @typescript-eslint/no-explicit-any

/home/claude-testing/vikunja/frontend/src/views/project/helpers/useGanttTaskList.ts
  44:114  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any

/home/claude-testing/vikunja/frontend/src/views/project/settings/ProjectSettingsBackground.vue
  209:23  error  Unexpected any. Specify a different type                                                       @typescript-eslint/no-explicit-any
  210:8   error  'backgroundResponse' is assigned a value but never used. Allowed unused vars must match /^_/u  @typescript-eslint/no-unused-vars

/home/claude-testing/vikunja/frontend/src/views/project/settings/ProjectSettingsWebhooks.vue
  205:7  warning  v-on event '@update:model-value' can't be hyphenated  vue/v-on-event-hyphenation

/home/claude-testing/vikunja/frontend/src/views/sharing/LinkSharingAuth.vue
  134:15  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any

/home/claude-testing/vikunja/frontend/src/views/tasks/ShowTasks.vue
  14:1   error    Expected indentation of 4 tabs but found 3 tabs                         vue/html-indent
  15:1   error    Expected indentation of 4 tabs but found 3 tabs                         vue/html-indent
  15:32  warning  Expected 1 line break before closing bracket, but no line breaks found  vue/html-closing-bracket-newline

/home/claude-testing/vikunja/frontend/src/views/tasks/TaskDetailView.vue
  724:68  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  794:61  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  798:15  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  879:39  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  880:26  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any

/home/claude-testing/vikunja/frontend/src/views/teams/EditTeam.vue
  108:22  error    Filters are deprecated                                                       vue/no-deprecated-filter
  115:8   warning  v-on event '@update:model-value' can't be hyphenated                         vue/v-on-event-hyphenation
  287:14  error    'IAbstract' is defined but never used. Allowed unused vars must match /^_/u  @typescript-eslint/no-unused-vars

/home/claude-testing/vikunja/frontend/src/views/user/Login.vue
  241:14  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any

/home/claude-testing/vikunja/frontend/src/views/user/PasswordReset.vue
  91:37  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  92:14  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any

/home/claude-testing/vikunja/frontend/src/views/user/Register.vue
  203:14  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any

/home/claude-testing/vikunja/frontend/src/views/user/RequestPasswordReset.vue
  83:14  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any

/home/claude-testing/vikunja/frontend/src/views/user/settings/ApiTokens.vue
  21:44  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  25:48  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  26:53  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  53:40  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  70:40  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  97:61  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  98:51  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any

/home/claude-testing/vikunja/frontend/src/views/user/settings/Avatar.vue
  124:86  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any

/home/claude-testing/vikunja/frontend/src/views/user/settings/General.vue
  535:105  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  541:151  error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  546:73   error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any
  546:87   error  Unexpected any. Specify a different type  @typescript-eslint/no-explicit-any

✖ 246 problems (242 errors, 4 warnings)
  19 errors and 4 warnings potentially fixable with the `--fix` option.

 ELIFECYCLE  Command failed with exit code 1.
