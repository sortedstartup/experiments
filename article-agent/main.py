from agents import Agent, Runner
from agents.mcp import MCPServerSse, MCPServerStdio
import asyncio
import os


async def main():
    # Playwright MCP Server (SSE connection to running container)
    async with MCPServerSse(
        name="Playwright Server",
        params={
            "url": "http://localhost:8931/sse",
        },
    ) as playwright_server:
        # Brave Search MCP Server (stdio)
        async with MCPServerStdio(
            name="Brave Search",
            params={
                "command": "npx",
                "args": ["-y", "@brave/brave-search-mcp-server"],
                "env": {
                    "BRAVE_API_KEY": os.getenv("BRAVE_API_KEY")
                },
            },
        ) as brave_server:
            agent = Agent(
                name="Web Agent",
                instructions="""You are an **Article Creator Agent**. 
                Your main job is to take a topic and create an article about it.

                You can use the following tools:
                - Brave Search: To search for information on the web.
                - Playwright: To scrape the web.

                Steps to follow for creating an article from the users requirements:
                1. **Search:**
                    - Use Brave Search to search for information on the web use TOP 2 results.
                2. **Scrape:**
                    - Use Playwright to scrape the web.
                3. **Write:**
                    - Use the information from the web to write an article.
                """,
                mcp_servers=[playwright_server, brave_server],
                model="gpt-5-mini-2025-08-07",
            )
            
            result = await Runner.run(
                agent,
                "Latest news in india"
            )
            print(result.final_output)


if __name__ == "__main__":
    asyncio.run(main())
