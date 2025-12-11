from agents import Agent, Runner, function_tool, AgentHooks
import asyncio
import subprocess
import os
import re

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
            f"mkdir -p {target_path} && cd {target_path} && sudo git clone --depth 1 {repo_url} && cd Go_gPRC_Template_Repo/ && sudo rm -rf .git"
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
    """Read a file from inside the Docker container."""
    print("Tool: Read file (container) -> " + file_path)

    try:
        cmd = [
            "docker", "exec", "work-dev-1",
            "bash", "-c",
            f"sudo cat {file_path}"
        ]

        result = subprocess.run(
            cmd,
            capture_output=True,
            text=True,
            timeout=10
        )

        if result.returncode == 0:
            return {
                "status": "success",
                "content": result.stdout,
                "message": f"Read file {file_path} successfully."
            }
        else:
            return {
                "status": "error",
                "message": result.stderr.strip()
            }

    except Exception as e:
        return {
            "status": "error",
            "message": f"Error reading file {file_path}: {str(e)}"
        }



@function_tool
def write_file(file_path: str, content: str) -> dict:
    """Write a file inside the Docker container with proper permissions."""
    print("Tool: Write file (container) -> " + file_path)

    try:
        # Escape single quotes for shell
        escaped = content.replace("'", "'\"'\"'")

        # Build command: create dir as root, write file with sudo tee
        cmd = (
            f"sudo mkdir -p $(dirname {file_path}) && "
            f"echo '{escaped}' | sudo tee {file_path} > /dev/null && "
            f"sudo chmod 664 {file_path}"
        )

        result = subprocess.run(
            ["docker", "exec", "work-dev-1", "bash", "-c", cmd],
            capture_output=True,
            text=True,
            timeout=10
        )

        if result.returncode == 0:
            return {
                "status": "success",
                "message": f"Wrote content to file {file_path} successfully."
            }
        else:
            return {
                "status": "error",
                "message": result.stderr.strip()
            }

    except Exception as e:
        return {
            "status": "error",
            "message": f"Failed writing file {file_path}: {str(e)}"
        }




@function_tool
def grep_file(file_path: str, pattern: str) -> dict:
    """Search for a regex pattern inside a file inside the Docker container."""
    print("Tool: Grep file (container) -> " + file_path)

    try:
        # Escape single quotes in the pattern for shell safety
        safe_pattern = pattern.replace("'", "'\"'\"'")

        cmd = [
            "docker", "exec", "work-dev-1",
            "bash", "-c",
            f"sudo grep -E '{safe_pattern}' {file_path} || true"
        ]

        result = subprocess.run(
            cmd,
            capture_output=True,
            text=True,
            timeout=10
        )

        if result.returncode == 0 or result.returncode == 1:
            # grep exit code 1 means "no matches", not an error
            lines = result.stdout.strip().split("\n") if result.stdout else []

            matches = [line for line in lines if line.strip()]

            return {
                "status": "success",
                "matches": matches,
                "message": f"Found {len(matches)} matches."
            }

        return {
            "status": "error",
            "message": result.stderr.strip()
        }

    except Exception as e:
        return {
            "status": "error",
            "message": f"Error grepping {file_path}: {str(e)}"
        }


@function_tool
def run_template_runner(json_data: str, directory: str) -> dict:
    """Run the template runner with the given json and directory."""
    print("Tool: Run template runner -> " + json_data + " " + directory)

    try:

        cmd = [
            "docker", "exec", "work-dev-1",
            "bash", "-c",
            f"cd {directory} && sudo /sorted/template-runner -json '{json_data}' -dir {directory}"
        ]

        result = subprocess.run(
            cmd,
            capture_output=True,
            text=True,
            timeout=120
        )

        if result.returncode == 0:
            return {
                "status": "success",
                "message": "Template runner completed successfully."
            }
        else:
            return {
                "status": "error",
                "message": result.stderr
            }
    except Exception as e:
        return {
            "status": "error",
            "message": f"Error running template runner: {str(e)}"
        }

@function_tool
def init_git_repo(directory: str, user_email: str="sanskaraggarwal2025@gmail.com", user_name: str="sanskaraggarwal") -> dict:
    """Initialize a git repository inside the Docker container with the given user email and user name."""
    print("Tool: Init git repo -> " + directory)
    try:

        cmd = [
            "docker", "exec", "work-dev-1",
            "bash", "-c",
            f"cd {directory} && sudo git init"
        ]

        result = subprocess.run(
            cmd,
            capture_output=True,
            text=True,
            timeout=10
        )
        if result.returncode == 0:
            subprocess.run([
                "docker", "exec", "work-dev-1",
                "sudo", "git", "-C", directory, "config", "user.email", user_email
            ], check=True)

            subprocess.run([
                "docker", "exec", "work-dev-1",
                "sudo", "git", "-C", directory, "config", "user.name", user_name
            ], check=True)

            return {
                "status": "success",
                "message": "Git repository initialized successfully."
            }
        else:
            return {
                "status": "error",
                "message": result.stderr
            }
    except Exception as e:
        return {
            "status": "error",
            "message": f"Error initializing git repository: {str(e)}"
        }

@function_tool
def commit_git_repo(directory: str, message: str) -> dict:
    """Commit a git repository inside the Docker container."""
    print("Tool: Commit git repo -> " + directory + " " + message)
    try:
        cmd = [
            "docker", "exec", "work-dev-1",
            "bash", "-c",
            f"cd {directory} && sudo git add . && sudo git commit -m '{message}'"
        ]
        result = subprocess.run(
            cmd,
            capture_output=True,
            text=True,
            timeout=10
        )
        if result.returncode == 0:
            return {
                "status": "success",
                "message": "Git repository committed successfully."
            }
        else:
            return {
                "status": "error",
                "message": result.stderr
            }
    except Exception as e:
        return {
            "status": "error",
            "message": f"Error committing git repository: {str(e)}"
        }


@function_tool
def autogenerate_proto_code(service_directory: str) -> dict:
    """Autogenerate the proto code for the given service directory inside the Docker container."""
    print("Tool: Autogenerate proto code -> " + service_directory)
    try:
        cmd = [
            "docker", "exec", "work-dev-1",
            "bash", "-c",
            f"cd {service_directory} && sudo /home/dev/sw/go/bin/go generate"
        ]
        result = subprocess.run(
            cmd,
            capture_output=True,
            text=True,
            timeout=10
        )
        if result.returncode == 0:
            return {
                "status": "success",
                "message": "Proto code autogenerated successfully."
            }
        else:
            return {
                "status": "error",
                "message": result.stderr
            }
    except Exception as e:
        return {
            "status": "error",
            "message": f"Error autogenerating proto code: {str(e)}"
        }

class CustomAgentHooks(AgentHooks):
    async def on_tool_start(self, tool_context, agent, tool):
        print(f"Tool: {tool.name} started")

    async def on_tool_end(self, tool_context, agent, tool, result):
        print(f"Tool: {tool.name} completed")
        print(f"Tool: {tool.name} output: {result}")


async def main():
    agent = Agent(
        name="Zero to Release Agent",
        instructions="""
You are a Zero to Release Agent.
Your job is to build a web app from template given in template repository.

<starter_template>
It is a go lang app with grpc proto files and service files

<file_structure>
    - /sorted/Go_gPRC_Template_Repo/backend/mono/main.go
    - /sorted/Go_gPRC_Template_Repo/backend/first_service/service.go --> add your APIs here
    - /sorted/Go_gPRC_Template_Repo/proto/service.proto --> proto file for the first service
    - /sorted/Go_gPRC_Template_Repo/backend/first_service/go.mod --> go mod file for the first service
    - /sorted/Go_gPRC_Template_Repo/backend/mono/go.mod --> go mod file for the mono service
    no database is used, use in memory structures to store the data
</file_structure>

</starter_template>

FOLLOW THESE STEPS STRICTLY:
1. First call connect_to_container.
2. Then call clone_repo_in_container with the repository url: https://github.com/sanskaraggarwal2025/Go_gPRC_Template_Repo.git and the target path: /sorted.
3. Read the Clone Repository project structure.
4. Inside each folder there are mod files, proto files and service files.
5. In each of these file there are template variable names like {{.Module}}, {{.ProjectModule}}, {{.ServiceModule}}.
6. Create a json for these template variables as per user requirement.
9. First call the /sorted/template-runner with appropriate json_data for proto/.
10. Then call the /sorted/template-runner with appropriate json_data for backend/.
11. After this call, the template files will be updated with the user requirement.
12. Initialize a git repository inside the /sorted/Go_gPRC_Template_Repo/ directory.
13. Based on changes done, commit the changes to the git repository with the appropriate message, after each logical step always commit the changes.
14. Based on user requirement, first determine the rpc required.
15. Based on rpc required, create the proto file.
16. Autogenerate the proto code using the /sorted/Go_gPRC_Template_Repo/backend/first_service/  go generate command.
16. Implement the rpc in the service file.
17. Once proto and service file are created, and in mono/main.go, create a grpc server and add the service to the server.
14. STOP!!

""",
        tools=[connect_to_container, clone_repo_in_container, read_file, write_file, grep_file, run_template_runner, init_git_repo, commit_git_repo, autogenerate_proto_code],
        model="gpt-5-mini-2025-08-07",
        hooks=CustomAgentHooks()
    )

    result = await Runner.run(
        agent,
        "Build a backend service for a Todo List application. Basic Template is here: https://github.com/sanskaraggarwal2025/Go_gPRC_Template_Repo.git. The service should be able to add, delete, update and get todos.Just save data in memory, no database is used.",
        max_turns=25
    )

    print(result.final_output)


if __name__ == "__main__":
    asyncio.run(main())
