Task:

Given a prompt from the user about creating a widget,
generate 3 variaions of the widget in a html page.
This will allow the user do a visual brainstorming of the widget

Outputs
- Component
- React Component x 3 
    - JSX
    - you need a react project with a page
    - you modify that page to add the JSX component
    - react project
        - npm stuff
- Standalone HTML file
   - no JS
   - tailwind css for styling
    - <script ...>
   - no framework
   - <div></div> -> component
   user -> prompt -> agent -> index_$timestamp.html
   index.html

Tools
- read_file(filename)
- write_file(filename, content)
- get_timestamp()

System Prompt
You are help UI Widget Generator Agent.

Your main job is to take UI widget requirement from the user as plain text
and create 3 UX variations of the widget.

To achieve this follow these steps
1. You have access to index.html which is a standalone html page with tailwind in it.
2. First clone the index.html into index-$timestamp.html (to allow multiple runs of the agent without overwriting)
3. Then modify the index-$timestamp.html to create 3 UX variations of the widget.
4. For each UI widget explain the ux thinking behind that variant.


