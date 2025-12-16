import { signIn } from "@/auth";

export default function Home() {
  return (
    <div className="flex min-h-screen items-center justify-center bg-zinc-50 font-sans dark:bg-black">
      <div className="flex flex-col items-center justify-center">
        <h1>Home</h1>
        <form
          action={async () => {
            "use server"
            await signIn("google")
          }}
        >
          <button type="submit">Signin with Google</button>
        </form>
      </div>
    </div>
  );
}
