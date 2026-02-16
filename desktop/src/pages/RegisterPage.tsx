import { useState, useEffect } from "react";
import { Link, useNavigate, useSearchParams } from "react-router-dom";
import { useForm } from "react-hook-form";
import { useAuthStore } from "../stores/authStore";
import type { RegisterRequest } from "../types/user";
import GoogleSignInButton from "../components/auth/GoogleSignInButton";
import { apiFetch } from "../services/api";

export default function RegisterPage() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const inviteToken = searchParams.get("invite") || undefined;
  const { register: registerUser, error, clearError } = useAuthStore();
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [inviteValid, setInviteValid] = useState<boolean | null>(null);
  const { register, handleSubmit, formState: { errors } } = useForm<RegisterRequest>();

  // Validate invite token on mount
  useEffect(() => {
    if (inviteToken) {
      apiFetch(`/invitations/validate/${inviteToken}`)
        .then(() => setInviteValid(true))
        .catch(() => setInviteValid(false));
    }
  }, [inviteToken]);

  const onSubmit = async (data: RegisterRequest) => {
    setIsSubmitting(true);
    clearError();
    try {
      await registerUser({
        ...data,
        invite_token: inviteToken,
      });
      navigate("/");
    } catch {
      // error is set in store
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="flex h-full items-center justify-center bg-surface">
      <div className="w-full max-w-sm px-4 py-6 sm:p-8">
        <div className="mb-6 flex justify-center">
          <img src="/feather-logo.png" alt="Feather" className="h-16 w-16 rounded-full" />
        </div>

        <h1 className="mb-8 text-center text-2xl font-bold text-gray-900 dark:text-gray-100">
          Create Account
        </h1>

        {inviteToken && inviteValid === false && (
          <div className="mb-4 rounded bg-yellow-50 p-3 text-sm text-yellow-700 dark:bg-yellow-900/20 dark:text-yellow-400">
            This invite link is invalid or has expired.
          </div>
        )}

        {inviteToken && inviteValid === true && (
          <div className="mb-4 rounded bg-green-50 p-3 text-sm text-green-700 dark:bg-green-900/20 dark:text-green-400">
            You've been invited to join Feather!
          </div>
        )}

        {error && (
          <div className="mb-4 rounded bg-red-50 p-3 text-sm text-red-600 dark:bg-red-900/20 dark:text-red-400">
            {error}
          </div>
        )}

        <GoogleSignInButton />

        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
          <div>
            <label className="mb-1 block text-sm font-medium text-gray-700 dark:text-gray-300">
              Name
            </label>
            <input
              type="text"
              {...register("name", { required: "Name is required", minLength: { value: 2, message: "Min 2 characters" } })}
              className="w-full rounded border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none dark:border-gray-600 dark:bg-surface-secondary dark:text-gray-100"
              placeholder="Your name"
            />
            {errors.name && (
              <p className="mt-1 text-xs text-red-500">{errors.name.message}</p>
            )}
          </div>

          <div>
            <label className="mb-1 block text-sm font-medium text-gray-700 dark:text-gray-300">
              Email
            </label>
            <input
              type="email"
              {...register("email", { required: "Email is required" })}
              className="w-full rounded border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none dark:border-gray-600 dark:bg-surface-secondary dark:text-gray-100"
              placeholder="you@example.com"
            />
            {errors.email && (
              <p className="mt-1 text-xs text-red-500">{errors.email.message}</p>
            )}
          </div>

          <div>
            <label className="mb-1 block text-sm font-medium text-gray-700 dark:text-gray-300">
              Password
            </label>
            <input
              type="password"
              {...register("password", { required: "Password is required", minLength: { value: 8, message: "Min 8 characters" } })}
              className="w-full rounded border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none dark:border-gray-600 dark:bg-surface-secondary dark:text-gray-100"
              placeholder="Min 8 characters"
            />
            {errors.password && (
              <p className="mt-1 text-xs text-red-500">{errors.password.message}</p>
            )}
          </div>

          <button
            type="submit"
            disabled={isSubmitting}
            className="w-full rounded bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-50"
          >
            {isSubmitting ? "Creating account..." : "Create Account"}
          </button>
        </form>

        <p className="mt-4 text-center text-sm text-gray-500 dark:text-gray-400">
          Already have an account?{" "}
          <Link to="/login" className="text-blue-600 hover:underline">
            Sign In
          </Link>
        </p>
      </div>
    </div>
  );
}
