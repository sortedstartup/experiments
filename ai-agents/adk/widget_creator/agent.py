from google.adk.agents.llm_agent import Agent

import datetime
import os

def read_file(filename: str) -> str:
    """Reads the content of a file."""
    with open(filename, 'r') as f:
        return f.read()

def write_file(filename: str, content: str) -> str:
    """Writes content to a file."""
    with open(filename, 'w') as f:
        f.write(content)
    return f"Successfully wrote to {filename}"

def get_timestamp() -> str:
    """Returns the current timestamp."""
    return datetime.datetime.now().strftime("%Y%m%d%H%M%S")

root_agent = Agent(
    model='gemini-3-flash-preview',
    name='widget_creator',
    description='You are a widget creator agent. ',
    instruction="""
        You are help UI Widget Generator Agent.

        Your main job is to take UI widget requirement from the user as plain text
        and create 3 UX variations of the widget.

        To achieve this follow these steps
        1. You have access to index.html which is a standalone html page with tailwind in it.
        2. First clone the index.html into index-$timestamp.html (to allow multiple runs of the agent without overwriting)
        3. Then modify the index-$timestamp.html to create 5 UX variations of the widget.
        4. Make sure you have all the variants in the same file.
        5. For each UI widget explain the ux thinking behind that variant.

        if you need logos/ icon use this online service from google in a image tag
        <img src="https://www.google.com/s2/favicons?domain=github.com&sz=64">

        for general images use this service - https://picsum.photos/400/300, where 400 and 300 is width and height
    """,
    tools=[read_file, write_file, get_timestamp],
)
