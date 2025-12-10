from agents import Agent, Runner, function_tool, AgentHooks
import asyncio
import subprocess


@function_tool
def connect_to_container() -> str:
    """Check that the container is running and reachable."""
    try:
        result = subprocess.run(
            ["docker", "exec", "work-dev-1", "echo", "connected"],
            capture_output=True,
            text=True,
            timeout=10
        )
        if result.returncode == 0:
            return "Container is reachable."
        return f"Failed to reach container:\n{result.stderr}"
    except subprocess.TimeoutExpired:
        return "Timeout while checking container."
    except Exception as e:
        return f"Error: {e}"


@function_tool
def clone_repo_in_container(repo_url: str, target_path: str) -> str:
    """Clone a git repository inside the Docker container."""
    try:
        cmd = [
            "docker", "exec", "work-dev-1",
            "bash", "-c",
            f"cd {target_path} && sudo git clone {repo_url}"
        ]

        result = subprocess.run(
            cmd,
            capture_output=True,
            text=True,
            timeout=120
        )

        if result.returncode == 0:
            return f"Repository cloned successfully:\n{result.stdout}"
        else:
            return f"Clone failed:\n{result.stderr}"

    except subprocess.TimeoutExpired:
        return "Clone operation timed out."
    except Exception as e:
        return f"Error: {e}"


class CustomAgentHooks(AgentHooks):
    async def on_tool_start(self, context, agent, tool):
        print("Tool: " + tool.name + " started")


async def main():
    agent = Agent(
        name="Zero to Release Agent",
        instructions="""
You are a Clone Agent.
Your job is to clone a GitHub repository inside a running Docker container.

TOOLS:
- connect_to_container: Verify container is reachable.
- clone_repo_in_container: Clone a GitHub repo to a directory inside container.

STEPS:
1. First call connect_to_container.
2. Then call clone_repo_in_container with:
   repo: https://github.com/sanskaraggarwal2025/Go_gPRC_Template_Repo.git
   target path: /usr/local/
""",
        tools=[connect_to_container, clone_repo_in_container],
        model="gpt-5-mini-2025-08-07",
        hooks=CustomAgentHooks()
    )

    result = await Runner.run(
        agent,
        "Clone the GitHub repo https://github.com/sanskaraggarwal2025/Go_gPRC_Template_Repo.git into /usr/local/ inside the container."
    )

    print(result.final_output)

if __name__ == "__main__":
    asyncio.run(main())
