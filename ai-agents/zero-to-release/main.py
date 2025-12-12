from agents import Agent, Runner, function_tool, AgentHooks, SQLiteSession
import asyncio
import subprocess
import os
import re
import tempfile
import uuid

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
            f"mkdir -p {target_path} && cd {target_path} && git clone --depth 1 {repo_url} && cd Go_gPRC_Template_Repo/ && rm -rf .git"
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
            f"cat {file_path}"
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
            f"mkdir -p $(dirname {file_path}) && "
            f"echo '{escaped}' |  tee {file_path} > /dev/null && "
            f"chmod 664 {file_path}"
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
            f"grep -E '{safe_pattern}' {file_path} || true"
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
            f"cd {directory} && /home/dev/sorted/template-runner -json '{json_data}' -dir {directory}"
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
            f"cd {directory} && git init"
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
                "git", "-C", directory, "config", "user.email", user_email
            ], check=True)

            subprocess.run([
                "docker", "exec", "work-dev-1",
                "git", "-C", directory, "config", "user.name", user_name
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
            f"cd {directory} && git add . && git commit -m '{message}'"
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
    """Autogenerate the proto code for the given service directory inside the Docker container.
    Args:
        service_directory (str): Directory of the service where autogenerate.go is present.
    Returns:
        dict: Result of the autogeneration.
    """
    print("Tool: Autogenerate proto code -> " + service_directory)
    try:
        cmd = [
            "docker", "exec", "work-dev-1",
            "bash", "-c",
            f"cd {service_directory} && go generate"
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


@function_tool
def user_review_tool(content: str) -> str:
    """
    Take review from user for the given content.

    Args:
        content (str): Content to be reviewed.

    Returns:
        str: Reviewed content.
    """
    print(f"Tool: User Review")
    try:
        with tempfile.NamedTemporaryFile(mode='w', delete=False, suffix=".txt") as tmp:
            tmp.write(content)
            tmp_path = tmp.name

        try:
            # Open gedit and wait
            subprocess.run(["gedit", "--wait", tmp_path], check=True)
            
            # Read back
            with open(tmp_path, "r") as f:
                reviewed_content = f.read()
        finally:
            if os.path.exists(tmp_path):
                os.remove(tmp_path)
            
        return reviewed_content

    except Exception as e:
        return f"Error during user review: {e}"


@function_tool
def build_app_tool(directory: str) -> str:
    """ 
    Tool to build application and see if there is any error
    Args:
        directory (str): directory where we run go build
    Returns:
        str: build output success/failure
    """

    try:
        cmd = [
            "docker", "exec", "work-dev-1",
            "bash", "-c",
            f"cd {directory} && go build"
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
                "message": "Application built successfully."
            }
        else:
            return {
                "status": "error",
                "message": result.stderr
            }
    except Exception as e:
        return {
            "status": "error",
            "message": f"Error building application: {str(e)}"
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
It is a go lang app with grpc proto, api layer and database files (dao layer) using sqlite and sqlx.
It has one working rpc from proto to dao layer just for reference, take Reference from it and start building user required application by editing following files.

<file_structure>
    - /home/dev/sorted/Go_gPRC_Template_Repo/backend/mono/main.go
    - /home/dev/sorted/Go_gPRC_Template_Repo/backend/first_service/api/api.go --> add your APIs here
    - /home/dev/sorted/Go_gPRC_Template_Repo/proto/service.proto --> proto file for the first service
    - /home/dev/sorted/Go_gPRC_Template_Repo/backend/first_service/go.mod --> go mod file for the first service
    - /home/dev/sorted/Go_gPRC_Template_Repo/backend/mono/go.mod --> go mod file for the mono service
    - /home/dev/sorted/Go_gPRC_Template_Repo/backend/first_service/dao/dao.go --> interface for the database operations
    - /home/dev/sorted/Go_gPRC_Template_Repo/backend/first_service/dao/dao_sqlite.go --> implementation of the database operations using sqlite and sqlx
    - /home/dev/sorted/Go_gPRC_Template_Repo/backend/first_service/dao/migrate.go --> migration file for the database
    - /home/dev/sorted/Go_gPRC_Template_Repo/backend/first_service/dao/db/migrations/1_init.up.sql -> it has all the migration files for the database (UP) scripts for the database
    - /home/dev/sorted/Go_gPRC_Template_Repo/backend/first_service/dao/models.go --> it has the models for the database
</file_structure>

</starter_template>

GUIDELINES
- As you go on working document important decisions you are making by appending them to file decision_log.md
- each decision should have a timestamp and should be in technical precise language in bullet points
- each decision should contain reasoning behind the decision taken.
- after each file generation, take review from user for the full file changes using user_review_tool tool.
- user will reivew the changes in this format //REVIEW: <review>, you have to make those changes and proceed further.
- take commit after each logical step.


FOLLOW THESE STEPS STRICTLY:
1. First call connect_to_container.
2. Then call clone_repo_in_container with the repository url: https://github.com/sanskaraggarwal2025/Go_gPRC_Template_Repo.git and the target path: /home/dev/sorted.
3. Read the Clone Repository project structure.
5. Inside each folder there are mod files, proto files, api files and dao files.
6. In each of these file there are template variable names like {{.Module}}, {{.ProjectModule}}, {{.ServiceModule}}.
7. Create a json for these template variables as per user requirement.
8. First call the /home/dev/sorted/template-runner with appropriate json_data for proto/.
9. Then call the /home/dev/sorted/template-runner with appropriate json_data for backend/.
10. After this call, the template files will be updated with the user requirement.
12. Initialize a git repository inside the /home/dev/sorted/Go_gPRC_Template_Repo/ directory.
13. After each file generation, use this TOOL user_review_tool to**take review from user for the full file changes**.
14. User will reivew the changes in this format //REVIEW: <review>, you have to make those changes and proceed further.
15. Based on changes done, commit the changes to the git repository with the appropriate message, after each logical step always commit the changes.
16. Based on **user requirement**, first determine the **rpc required, api layer and database layer changes required**.
17. Based on rpc required, create the proto file.
18. After every proto file change, autogenerate the proto code using autogenerate_proto_code tool.
19. Based on user review, make the changes and proceed further.
20. Edit proto file, first_service/api/api.go for api changes, first_service/dao/dao.go for dao interfaces changes first_service/dao/dao_sqlite.go for database implementation changes.
20. Once proto is done proceed with the api layer and database layer changes.
21. Once proto and api layer and database layer are created.
22. **Build the application using **build_app_tool** tool and see if it is working or not**.
23. If it is not working, make the changes and proceed further.
24. If it is working, create implementation_proto.md file where you have to write the implementation details at proto level like assumptions, architectural_decisions, decisions taken in bullet points.
25. After you are done with all implementation, create implementation_dao.md file where you have to write the implementation details at DAO level like assumptions, architectural_decisions, decisions taken in bullet points.
26. STOP!!

""",
        tools=[connect_to_container, clone_repo_in_container, read_file, write_file, grep_file, run_template_runner, init_git_repo, commit_git_repo, autogenerate_proto_code, user_review_tool, build_app_tool],
        model="gpt-5-mini-2025-08-07",
        hooks=CustomAgentHooks()
    )

    uid = uuid.uuid4()
    str_uid = str(uid)
    session = SQLiteSession(str_uid, "conversations.db")

    result = await Runner.run(
        agent,
        "Build a backend service for a Todo List application. Basic Template is here: https://github.com/sanskaraggarwal2025/Go_gPRC_Template_Repo.git. Create a todo list application with grpc proto, api layer and database files (dao layer) using sqlite and sqlx.",
        max_turns=75,
        # session=session
    )
    usage = result.context_wrapper.usage
    print("Requests:", usage.requests)
    print("Input tokens:", usage.input_tokens)
    print("Output tokens:", usage.output_tokens)
    print("Cached tokens:", usage.input_tokens_details.cached_tokens)
    print("Total tokens:", usage.total_tokens)

    print(result.final_output)


if __name__ == "__main__":
    asyncio.run(main())
