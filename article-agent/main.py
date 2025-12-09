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
                instructions="You provide assistance with web scraping and search queries. Use Playwright for navigating websites and extracting data. Use Brave Search for searching the web.",
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
