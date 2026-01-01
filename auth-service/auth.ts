import { betterAuth } from "better-auth";

const auth = betterAuth({
    secret: process.env.BETTER_AUTH_SECRET,
    url: process.env.BETTER_AUTH_URL,
});

export default auth;