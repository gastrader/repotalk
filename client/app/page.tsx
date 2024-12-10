"use client";
import { Loader } from "lucide-react";
import Image from "next/image";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useEffect, useRef, useState } from "react";

export default function Home() {
  const inputRef = useRef<HTMLInputElement>(null);
  const [isDisabled, setIsDisabled] = useState(false);
  const router = useRouter();

  useEffect(() => {
    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === "/") {
        event.preventDefault();
        inputRef.current?.focus();
      }
    };

    document.addEventListener("keydown", handleKeyDown);

    return () => {
      document.removeEventListener("keydown", handleKeyDown);
    };
  }, []);

  const handleSubmit = async (event: React.FormEvent) => {
    setIsDisabled(true);
    event.preventDefault();
    const formData = new FormData(event.target as HTMLFormElement);
    const githubUrl = formData.get("githubUrl");

    const response = await fetch("http://localhost:8080/api/v1/crawl", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ githubUrl }),
    });

    if (response.ok) {
      const data = await response.json(); // Parse JSON response
      const { username, reponame, threadID } = data;
      setIsDisabled(false);
      router.push(`/${username}/${reponame}?tid=${threadID}`);
    } else {
      setIsDisabled(false);
      console.error("Error fetching repo");
    }
  };
  return (
    <div className="grid grid-rows-[20px_1fr_20px] items-center justify-items-center min-h-screen p-8 pb-20 gap-16 sm:p-20 font-[family-name:var(--font-geist-sans)]">
      <main className="flex flex-col row-start-2 items-center sm:items-start w-full max-w-md">
        <div className=" flex items-center justify-center w-full">
          <Image
            src="/sponge-3.webp"
            alt="logo"
            className="w-48"
            width={0}
            height={0}
            sizes="100vw"
            style={{ width: "40%", height: "auto" }}
            priority
          />

          <h1 className="align-middle font-[family-name:var(--font-tiny5)] text-2xl font-bold uppercase text-white md:text-4xl">
            REPO<span className="text-[#b2b937]">TALK</span>
          </h1>
        </div>
        <form className="w-full max-w-md space-y-4" onSubmit={handleSubmit}>
          <div className="relative w-full">
            <input
              ref={inputRef}
              name="githubUrl"
              className="flex h-10 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50 w-full rounded-lg border-2 border-[#595e00] bg-[#242600] px-4 py-2 pl-10 pr-20 text-white transition-all duration-300 focus:border-[#b2b937] focus:outline-none focus:ring-2 focus:ring-[#b2b937]"
              placeholder="enter github url..."
              type="text"
            />
            <svg
              width="24"
              height="24"
              stroke="currentColor"
              strokeWidth="2"
              className="lucide-icon lucide lucide-search absolute left-3 top-1/2 -translate-y-1/2 transform text-gray-500"
              xmlns="http://www.w3.org/2000/svg"
              viewBox="0 0 24 24"
              fill="none"
              strokeLinecap="round"
              strokeLinejoin="round"
            >
              <circle cx="11" cy="11" r="8"></circle>
              <path d="m21 21-4.3-4.3"></path>
            </svg>{" "}
            <div className="absolute right-3 top-1/2 -translate-y-1/2 transform text-sm text-gray-500 sm:inline hidden">
              <kbd className="rounded-sm border border-gray-600 bg-gray-600 px-1 py-0.5 text-xs font-semibold text-gray-900">
                /
              </kbd>
              <span className="ml-1">to focus</span>
            </div>
          </div>{" "}
          <div className="flex items-center justify-center">
            <button
              type="submit"
              tabIndex={0}
              className="ring-offset-background focus-visible:ring-ring inline-flex items-center justify-center whitespace-nowrap text-sm font-medium focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 h-10 py-2 rounded-lg bg-[#878c29] px-8 text-white transition-colors duration-300 hover:bg-[#333601]"
              data-button-root=""
              disabled={isDisabled}
            >
              {isDisabled ? (
                <Loader className="animate-spin h-4 w-4" />
              ) : (
                "Search"
              )}
            </button>
          </div>
        </form>
      </main>
      <footer className="row-start-3 flex flex-col items-center justify-center text-center">
        <span className="focus:ring-ring inline-flex select-none items-center rounded-full border px-2.5 py-0.5 text-xs font-semibold transition-colors focus:outline-none focus:ring-2 focus:ring-offset-2 mb-2 border-[#b2b937] text-[#b2b937]">
          63 Repos Indexed
        </span>

        <div className="flex items-center gap-2 ">
          <p className=" font-light">
            made by{" "}
            <Link
              className="text-[#b2b937] pr-1 hover:underline hover:underline-offset-4"
              target="_blank"
              rel="noopener noreferrer"
              href="https://plumega.com"
            >
              GP
            </Link>
            or{" "}
            <Link
              href="https://x.com/gastrading"
              className="text-[#b2b937] hover:underline hover:underline-offset-4 "
              target="_blank"
              rel="noopener noreferrer"
            >
              @gastrading
            </Link>{" "}
            <span className="text-gray-600">
              <em>on [x] dot com</em>
            </span>
          </p>
        </div>
      </footer>
    </div>
  );
}
