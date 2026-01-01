import { Elysia } from "elysia";
import { auth } from "../auth";

const app = new Elysia()
  .mount(auth.handler)
  .onRequest(({ request }) => {
    console.log(
      `[${new Date().toISOString()}] ${request.method} ${request.url}`,
    );
  })
  .onError(({ request, error }) => {
    console.group(
      `[${new Date().toISOString()}] ${request.method} ${request.url}`,
    );
    console.log(`ERROR: ${String(error)}`);
    console.groupEnd();
  })
  .listen(3000);

console.log(
  `🦊 Elysia is running at ${app.server?.hostname}:${app.server?.port}`
);
