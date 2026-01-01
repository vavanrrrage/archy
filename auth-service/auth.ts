import { betterAuth } from "better-auth";
import { jwt } from "better-auth/plugins";
import { prismaAdapter } from "better-auth/adapters/prisma";
import { PrismaClient } from "./prisma/generated/prisma/client";
import { PrismaPg } from "@prisma/adapter-pg";
import { transporter } from "./src/mail/transporter";

const adapter = new PrismaPg({ 
    connectionString: process.env.DATABASE_URL 
  });

const prisma = new PrismaClient({
    adapter
});

export const auth = betterAuth({
    secret: process.env.BETTER_AUTH_SECRET,
    url: process.env.BETTER_AUTH_URL,
    emailAndPassword: { 
        enabled: true, 
        requireEmailVerification: true
      },
    database: prismaAdapter(prisma, {
        provider: 'postgresql',
    }),
    plugins: [
        jwt()
    ],
    emailVerification: {
        sendVerificationEmail: async ({ user, url, token }, request) => {
            if (process.env.NODE_ENV !== "production") {
                console.log("sendVerificationEmail", user, url, token, request);
            } else {
                try {
                    console.log("Sending email to ", user.email);
                    await transporter.sendMail({
                      from: process.env.MAIL_FROM,
                      to: user.email,
                      subject: process.env.MAIL_SUBJECT_VERIFICATION,
                      text: url,
                      html: url,
                    });
                  } catch (error) {
                    console.error("Email is not sent - error: ", error);
                  }
            }
        },
        sendOnSignUp: true
    }
});
