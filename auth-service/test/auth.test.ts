import { describe, it, expect, beforeAll } from "bun:test";
import { Elysia } from "elysia";
import { auth } from "../auth";

const app = new Elysia().mount(auth.handler);

const baseURL = "http://localhost";

describe("Authentication API", () => {
  const testEmail = `test-${Date.now()}@example.com`;
  const testPassword = "TestPassword123!";
  const testName = "Test User";

  describe("POST /api/auth/sign-up/email", () => {
    it("should create a new user account", async () => {
      const response = await app.handle(
        new Request(`${baseURL}/api/auth/sign-up/email`, {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            email: testEmail,
            password: testPassword,
            name: testName,
          }),
        })
      );

      expect(response.status).toBe(200);
      const data = await response.json();
      expect(data.user).toBeDefined();
      expect(data.user.email).toBe(testEmail);
      expect(data.user.name).toBe(testName);
      expect(data.user.emailVerified).toBe(false);
    });

    it("should return error for duplicate email", async () => {
      const duplicateEmail = `duplicate-${Date.now()}@example.com`;

      await app.handle(
        new Request(`${baseURL}/api/auth/sign-up/email`, {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            email: duplicateEmail,
            password: testPassword,
            name: testName,
          }),
        })
      );

      const response = await app.handle(
        new Request(`${baseURL}/api/auth/sign-up/email`, {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            email: duplicateEmail,
            password: testPassword,
            name: testName,
          }),
        })
      );

      expect([400, 422]).toContain(response.status);
      const data = await response.json();
      expect(data.code).toBeDefined();
    });

    it("should return error for invalid email format", async () => {
      const response = await app.handle(
        new Request(`${baseURL}/api/auth/sign-up/email`, {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            email: "invalid-email",
            password: testPassword,
            name: testName,
          }),
        })
      );

      expect(response.status).toBe(400);
      const data = await response.json();
      expect(data.code).toBeDefined();
    });

    it("should return error for short password", async () => {
      const response = await app.handle(
        new Request(`${baseURL}/api/auth/sign-up/email`, {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            email: `short-pwd-${Date.now()}@example.com`,
            password: "short",
            name: testName,
          }),
        })
      );

      expect(response.status).toBe(400);
      const data = await response.json();
      expect(data.code).toBe("PASSWORD_TOO_SHORT");
    });
  });

  describe("POST /api/auth/sign-in/email", () => {
    let signedUpEmail: string;

    beforeAll(async () => {
      signedUpEmail = `signin-${Date.now()}@example.com`;
      await app.handle(
        new Request(`${baseURL}/api/auth/sign-up/email`, {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            email: signedUpEmail,
            password: testPassword,
            name: testName,
          }),
        })
      );
    });

    it("should return error for unverified email", async () => {
      const response = await app.handle(
        new Request(`${baseURL}/api/auth/sign-in/email`, {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            email: signedUpEmail,
            password: testPassword,
          }),
        })
      );

      expect([400, 403]).toContain(response.status);
      const data = await response.json();
      expect(data.code).toBe("EMAIL_NOT_VERIFIED");
    });

    it("should return error for incorrect password", async () => {
      const response = await app.handle(
        new Request(`${baseURL}/api/auth/sign-in/email`, {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            email: signedUpEmail,
            password: "WrongPassword123!",
          }),
        })
      );

      expect([400, 401]).toContain(response.status);
      const data = await response.json();
      expect(data.code || data.message).toBeDefined();
    });

    it("should return error for non-existent user", async () => {
      const response = await app.handle(
        new Request(`${baseURL}/api/auth/sign-in/email`, {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            email: `nonexistent-${Date.now()}@example.com`,
            password: testPassword,
          }),
        })
      );

      expect([400, 401]).toContain(response.status);
      const data = await response.json();
      expect(data.code || data.message).toBeDefined();
    });
  });

  describe("GET /api/auth/token", () => {
    it("should return 401 when not authenticated", async () => {
      const response = await app.handle(
        new Request(`${baseURL}/api/auth/token`, {
          method: "GET",
        })
      );

      expect(response.status).toBe(401);
    });

    it("should return 401 with invalid token", async () => {
      const response = await app.handle(
        new Request(`${baseURL}/api/auth/token`, {
          method: "GET",
          headers: {
            Cookie: "better-auth.session_token=invalid-token",
          },
        })
      );

      expect(response.status).toBe(401);
    });
  });
});

