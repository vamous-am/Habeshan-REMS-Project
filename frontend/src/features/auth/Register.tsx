import { zodResolver } from "@hookform/resolvers/zod";
import { useState } from "react";
import { useForm } from "react-hook-form";
import { Link, useNavigate } from "react-router-dom";
import { z } from "zod";
import {
  extractAuthErrorMessage,
  persistAuthSession,
  register,
} from "../../lib/api/authClient";
import {
  ETHIOPIAN_PHONE_MESSAGE,
  ETHIOPIAN_PHONE_PATTERN,
} from "../../lib/phone";
import { AuthField, AuthFieldGroup } from "./AuthField";
import { AuthButton, AuthError, AuthLayout } from "./AuthLayout";

const registerSchema = z.object({
  org_name: z.string().min(2, "Organization name must be at least 2 characters"),
  full_name: z.string().min(2, "Full name must be at least 2 characters"),
  email: z.string().email("Enter a valid email address"),
  password: z.string().min(8, "Password must be at least 8 characters"),
  phone: z
    .string()
    .regex(ETHIOPIAN_PHONE_PATTERN, ETHIOPIAN_PHONE_MESSAGE)
    .optional()
    .or(z.literal("")),
});

type RegisterForm = z.infer<typeof registerSchema>;

export default function Register() {
  const navigate = useNavigate();
  const [apiError, setApiError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  const form = useForm<RegisterForm>({
    resolver: zodResolver(registerSchema),
    defaultValues: {
      org_name: "",
      full_name: "",
      email: "",
      password: "",
      phone: "",
    },
  });

  async function handleSubmit(values: RegisterForm) {
    setLoading(true);
    setApiError(null);
    try {
      const session = await register({
        org_name: values.org_name.trim(),
        full_name: values.full_name.trim(),
        email: values.email.trim(),
        password: values.password,
        phone: values.phone?.trim() || undefined,
      });
      persistAuthSession(session);
      navigate("/tasks");
    } catch (err) {
      setApiError(extractAuthErrorMessage(err));
    } finally {
      setLoading(false);
    }
  }

  return (
    <AuthLayout
      title="Create your workspace"
      subtitle="Register your organization and your account to get started."
      footer={
        <p>
          Already have an account?{" "}
          <Link
            to="/login"
            className="font-semibold text-ochre-light underline-offset-2 hover:underline"
          >
            Sign in
          </Link>
        </p>
      }
    >
      {apiError && <AuthError message={apiError} />}

      <form
        className="flex flex-col gap-5"
        onSubmit={form.handleSubmit(handleSubmit)}
        noValidate
      >
        <AuthFieldGroup>
          <AuthField
            label="Organization"
            autoComplete="organization"
            placeholder="Agency name"
            error={form.formState.errors.org_name?.message}
            {...form.register("org_name")}
          />
          <AuthField
            label="Full name"
            autoComplete="name"
            placeholder="Your name"
            error={form.formState.errors.full_name?.message}
            {...form.register("full_name")}
          />
          <AuthField
            label="Email"
            type="email"
            autoComplete="email"
            placeholder="name@example.com"
            error={form.formState.errors.email?.message}
            {...form.register("email")}
          />
          <AuthField
            label="Password"
            type="password"
            autoComplete="new-password"
            placeholder="Password"
            passwordToggle
            hint="At least 8 characters"
            error={form.formState.errors.password?.message}
            {...form.register("password")}
          />
          <AuthField
            label="Phone"
            type="tel"
            autoComplete="tel"
            placeholder="09XXXXXXXX (optional)"
            error={form.formState.errors.phone?.message}
            {...form.register("phone")}
          />
        </AuthFieldGroup>
        <AuthButton type="submit" loading={loading} className="w-full">
          Create account
        </AuthButton>
      </form>
    </AuthLayout>
  );
}
