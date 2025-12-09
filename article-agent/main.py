from agents import Agent, Runner
from agents.mcp import MCPServerSse
import asyncio


async def main():
    async with MCPServerSse(
        name="Playwright Server",
        params={
            "url": "http://localhost:8931/sse",
        },
    ) as server:
        agent = Agent(
            name="Playwright Agent",
            instructions="You provide assistance with playwright queries. Get the data from the website.",
            mcp_servers=[server],
            model="gpt-5-mini-2025-08-07",
        )
        
        result = await Runner.run(
            agent,
            "Go to 'https://sortedstartup.com' and extract all <h1> text."
        )
        print(result.final_output)


if __name__ == "__main__":
    asyncio.run(main())
