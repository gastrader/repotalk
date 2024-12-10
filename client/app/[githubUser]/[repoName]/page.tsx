"use client";
import { Loader } from "lucide-react";
import Link from "next/link";
import { useParams, useRouter, useSearchParams } from "next/navigation";
import React, { useEffect, useRef, useState } from "react";

type QueryResponse = {
  message: string;
  username: string;
  reponame: string;
  threadID: string;
  response: string;
};

const RepoPage = () => {
  const inputRef = useRef<HTMLInputElement>(null);
  const scrollRef = useRef<HTMLDivElement>(null);
  const [isDisabled, setIsDisabled] = useState(false);
  const searchParams = useSearchParams();
  const router = useRouter()

  const tid = searchParams.get("tid");

  const params = useParams();
  const [messages, setMessages] = useState([
    { sender: "bot", text: "What would you like to know?" },
  ]);

  const { githubUser, repoName } = params;

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

  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
    }
  }, [messages]);

  const handleSubmit = async (event: React.FormEvent) => {
    setIsDisabled(true);
    event.preventDefault();
    const formData = new FormData(event.target as HTMLFormElement);
    const question = formData.get("question") as string | null;

    if (!question) {
      console.error("No question provided");
      setIsDisabled(false);
      return;
    }

    setMessages((prevMessages) => [
      ...prevMessages,
      {
        sender: "user",
        text: question,
      },
    ]);
    if (inputRef.current) {
      inputRef.current.value = "";
    }
    const response = await fetch("http://localhost:8080/api/v1/query", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ question, tid, githubUser, repoName }),
    });

    if (response.ok) {
      const data: QueryResponse = await response.json();

      if (!tid) {
        const newURL = `?tid=${data.threadID}`;
        router.push(newURL) // Update URL
      }

      setMessages((prevMessages) => [
        ...prevMessages,
        {
          sender: "bot",
          text: data.response,
        },
      ]);
      setIsDisabled(false);
    } else {
      setIsDisabled(false);
      console.error("Error fetching repo");
    }
  };
  return (
    <div className="grid grid-rows-[20px_1fr_20px] items-center justify-items-center min-h-screen p-8 pb-20 gap-16 sm:p-20 font-[family-name:var(--font-geist-sans)]">
      <header>
        <div className=" flex items-center justify-center w-full">
          <Link href="/">
            <h1 className="align-middle font-[family-name:var(--font-tiny5)] text-2xl font-bold uppercase text-white md:text-4xl">
              REPO<span className="text-[#b2b937]">TALK</span>
            </h1>
          </Link>
        </div>
      </header>
      <main className="flex flex-col row-start-2 items-end justify-center sm:items-end w-full max-w-lg xl:max-w-xl h-full  ">
        <span className="text-start font-mono w-full justify-start mb-2">
          talking with:{" "}
          <Link
            className="text-[#b2b937] hover:underline hover:underline-offset-4"
            href={`https://github.com/${githubUser}/${repoName}`}
            target="_blank"
            rel="noopener noreferrer"
          >
            github.com/{githubUser}/{repoName}
          </Link>
        </span>
        <div className="h-full w-full border-2 border-[#242600] rounded-md max-h-[500px] overflow-hidden">
          <div className="space-y-4 p-6 overflow-y-auto h-full" ref={scrollRef}>
            {messages.map((message, index) => (
              <div
                key={index}
                className={`flex flex-col ${
                  message.sender === "user" ? "items-end" : "items-start"
                }`}
              >
                <div
                  className={`p-3 rounded-lg max-w-xs relative ${
                    message.sender === "user"
                      ? "bg-[#797e25] text-white"
                      : "bg-[#242600] text-white"
                  }`}
                >
                  <div
                    className={`absolute ${
                      message.sender === "user"
                        ? "right-0 bg-[#797e25] translate-x-1 border-l-2 border-[#797e25] rotate-45"
                        : "left-0 bg-[#242600] -translate-x-1 border-r-2 border-[#242600] rotate-45"
                    } top-1/2 transform  -translate-y-1/2  w-3 h-3  border-2 `}
                  ></div>
                  {message.text}
                </div>
              </div>
            ))}
          </div>
        </div>
        <form
          className="w-full max-w-lg flex flex-row items-center justify-between space-x-4 mt-6 mx-auto"
          onSubmit={handleSubmit}
        >
          <div className="relative flex-1 w-full">
            <input
              ref={inputRef}
              name="question"
              className="w-full flex h-10 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50  rounded-lg border-2 border-[#595e00] bg-[#242600] px-4 py-2 text-white transition-all duration-300 focus:border-[#b2b937] focus:outline-none focus:ring-2 focus:ring-[#b2b937]"
              placeholder="Type here..."
              type="text"
            />
          </div>

          <div className="flex items-center justify-center">
            <button
              type="submit"
              tabIndex={0}
              className="ring-offset-background focus-visible:ring-ring inline-flex items-center justify-center whitespace-nowrap text-sm font-medium focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 h-10 rounded-lg bg-[#878c29] px-8 text-white transition-colors duration-300 hover:bg-[#333601]"
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
};

export default RepoPage;
