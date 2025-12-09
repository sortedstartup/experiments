from agents import Agent, Runner, function_tool, AgentHooks
from agents.mcp import MCPServerSse, MCPServerStdio
import asyncio
import os


@function_tool
def write_file(file_path: str, content: str) -> str:
    """
    Write content to a file at the specified path.
    
    Args:
        file_path: The path where the file should be written
        content: The content to write to the file
    
    Returns:
        A status message indicating success or failure
    """
    try:
        os.makedirs(os.path.dirname(file_path) if os.path.dirname(file_path) else ".", exist_ok=True)
        with open(file_path, 'w', encoding='utf-8') as f:
            f.write(content)
        return f"✅ Successfully wrote {len(content)} characters to {file_path}"
    except Exception as e:
        return f"❌ Error writing file: {str(e)}"


class CustomAgentHooks(AgentHooks):
    async def on_tool_start(self, context, agent, tool):
        print(" Tool: " + tool.name)


async def main():
    async with MCPServerSse(
        name="Playwright Server",
        params={"url": "http://localhost:8931/sse"},
    ) as playwright_server:
        async with MCPServerStdio(
            name="Brave Search",
            params={
                "command": "npx",
                "args": ["-y", "@brave/brave-search-mcp-server"],
                "env": {"BRAVE_API_KEY": os.getenv("BRAVE_API_KEY")},
            },
        ) as brave_server:
            agent = Agent(
                name="Web Agent",
                instructions="""You are an **Article Creator Agent**. 
                Your main job is to take a topic and create an article about it.

                You can use the following tools:
                - Brave Search: To search for information on the web.
                - Playwright: To scrape the web.
                - write_file: To save the article to a file.

                Steps to follow for creating an article from the users requirements:
                1. **Search:**
                    - Use Brave Search to search for information on the web use TOP 2 results.
                2. **Scrape:**
                    - Use Playwright to scrape the web.
                3. **Write:**
                    - Use write_file to save the article to a file (e.g., 'article.md').
                """,
                mcp_servers=[playwright_server, brave_server],
                tools=[write_file],
                model="gpt-5-mini-2025-08-07",
                hooks=CustomAgentHooks(),
            )
            
            result = await Runner.run(agent, "Latest news in india")
            print(result.final_output)


if __name__ == "__main__":
    asyncio.run(main())
