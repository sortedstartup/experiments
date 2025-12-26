import datetime
import os

from google.adk.agents import LlmAgent
from google.adk.agents.llm_agent import Agent
from google.adk.models.lite_llm import LiteLlm


def read_file(filename: str) -> str:
    """Reads the content of a file."""
    with open(filename, "r") as f:
        return f.read()


def write_file(filename: str, content: str) -> str:
    """Writes content to a file."""
    with open(filename, "w") as f:
        f.write(content)
    return f"Successfully wrote to {filename}"


def get_timestamp() -> str:
    """Returns the current timestamp."""
    return datetime.datetime.now().strftime("%Y%m%d%H%M%S")


# root_agent = LlmAgent(

# model=LiteLlm(model="openai/devstral-medium-latest"),
# model=LiteLlm(model="openai/mistral-large-latest"),
#    model=LiteLlm(model="openai/grok-4-1-fast-reasoning"),
#
root_agent = Agent(
    model="gemini-3-flash-preview",
    name="widget_creator",
    description="You are a widget creator agent. ",
    instruction="""
        You are a UI Widget Generator Agent.


        Your main job is to take UI widget requirement from the user as plain text
        and create UX design variations of the widget.


        To achieve this follow these steps
        1. You have access to index.html which is a standalone html page with tailwind in it.
        2. First clone the index.html into index-$timestamp.html (to allow multiple runs of the agent without overwriting)
        3. Then modify the index-$timestamp.html to create 5 UX design variations of the widget.
        4. Make sure you have all the variants in the same file.
        5. For each UI widget explain the ux thinking behind that variant.
        6. UX variant may chart and js library which are already present in index.html
        7. Add basic interactivity using jquery to give it a more real feel
        8. DO NOT Add any new library on your own

        if you need logos/ icon use this online service from google in a image tag
        <img src="https://www.google.com/s2/favicons?domain=github.com&sz=64">

        for general images use this service - https://picsum.photos/400/300, where 400 and 300 is width and height

        about index.html, here are the preadded things available, DO NOT add anything new
         - it has tailwind css
         - it has jquery for interactivve js
         - it has chart.js for charting
         - it has plotly another library for plotting
    """,
    tools=[read_file, write_file, get_timestamp],
)
