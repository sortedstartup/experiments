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


@function_tool
def read_file(file_path: str) -> dict:
    """Read a file and return its contents."""
    print("Tool: Read file -> " + file_path)

    try:
        with open(file_path, "r") as f:
            content = f.read()

        return {
            "status": "success",
            "content": content,
            "message": f"Read file {file_path} successfully."
        }

    except Exception as e:
        return {
            "status": "error",
            "message": f"Error reading file {file_path}: {str(e)}"
        }

@function_tool
def write_file(file_path: str, content: str) -> dict:
    """Write content to a file, creating parent dirs if needed."""
    print("Tool: Write file -> " + file_path)

    try:
        os.makedirs(os.path.dirname(file_path), exist_ok=True)

        with open(file_path, "w") as f:
            f.write(content)

        return {
            "status": "success",
            "message": f"Wrote content to file {file_path} successfully."
        }

    except Exception as e:
        return {
            "status": "error",
            "message": f"Failed writing file {file_path}: {str(e)}"
        }


@function_tool
def grep_file(file_path: str, pattern: str) -> dict:
    """Search for a regex pattern inside a file."""
    print("Tool: Grep file -> " + file_path)

    try:
        with open(file_path, "r") as f:
            lines = f.readlines()

        regex = re.compile(pattern)
        matches = [line.strip() for line in lines if regex.search(line)]

        return {
            "status": "success",
            "matches": matches,
            "message": f"Found {len(matches)} matches." if matches else "No matches found."
        }

    except re.error as err:
        return {
            "status": "error",
            "message": f"Invalid regex pattern: {err}"
        }

    except Exception as e:
        return {
            "status": "error",
            "message": f"Error reading {file_path}: {str(e)}"
        }


class CustomAgentHooks(AgentHooks):
    async def on_tool_start(self, context, agent, tool):
        print("Tool: " + tool.name + " started")


async def main():
    agent = Agent(
        name="Zero to Release Agent",
        instructions="""
You are a Zero to Release Agent.
Your job is to build a web app from template given in template repository.

<starter_template>
It is a go lang app with grpc proto files and service files

<file_structure>
	- backend/mono/main.go
	- backend/first_service/service.go --> add your APIs here
	- backend/proto/first_service.proto --> proto file for the first service
	no database is used, use in memory structures to store the data
</file_structure>

</starter_template>

STEPS:
1. First call connect_to_container.
2. Then call clone_repo_in_container with the repository url: https://github.com/sanskaraggarwal2025/Go_gPRC_Template_Repo.git and the target path: /usr/local/.
3. Read the Clone Repository project structure
4. As per the user requirement, name modules in go.mod file proto files and service files
5. In repository code, there are placeholders like {{.ModuleName}} for the module name, for the proto file name and {{.ProjectModule}} for the service name. Replace them with the user requirement.
6. Just replace the placeholders with the user requirement, no need to add any other code.
""",
        tools=[connect_to_container, clone_repo_in_container, read_file, write_file, grep_file],
        model="gpt-5-mini-2025-08-07",
        hooks=CustomAgentHooks()
    )

    result = await Runner.run(
        agent,
        "Build a backend service for a chat application. Basic Template is here: https://github.com/sanskaraggarwal2025/Go_gPRC_Template_Repo.git"
    )

    print(result.final_output)

if __name__ == "__main__":
    asyncio.run(main())
