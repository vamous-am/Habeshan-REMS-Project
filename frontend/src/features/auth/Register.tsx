import { zodResolver } from "@hookform/resolvers/zod";
import { useState } from "react";
import { useForm } from "react-hook-form";
import { Link, useNavigate } from "react-router-dom";
import { z } from "zod";
import { Button } from "../../components/Button";
import { Card } from "../../components/Card";
import { ErrorState } from "../../components/ErrorState";
import { Input } from "../../components/Input";
import {
  extractAuthErrorMessage,
  persistAuthSession,
  register,
} from "../../lib/api/authClient";

const registerSchema = z.object({
  org_name: z.string().min(2, "Organization name must be at least 2 characters"),
  full_name: z.string().min(2, "Full name must be at least 2 characters"),
  email: z.string().email("Enter a valid email address"),
  password: z.string().min(8, "Password must be at least 8 characters"),
  phone: z.string().max(20, "Phone number is too long").optional().or(z.literal("")),
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
    <div className="flex min-h-screen items-center justify-center bg-paper p-4">
      <Card
        title="Create account"
        subtitle="Register a new organization and employee account"
      >
        {apiError && (
          <div className="mb-4">
            <ErrorState message={apiError} />
          </div>
        )}

        <form
          className="flex flex-col gap-4"
          onSubmit={form.handleSubmit(handleSubmit)}
          noValidate
        >
          <Input
            label="Organization name"
            autoComplete="organization"
            error={form.formState.errors.org_name?.message}
            {...form.register("org_name")}
          />
          <Input
            label="Full name"
            autoComplete="name"
            error={form.formState.errors.full_name?.message}
            {...form.register("full_name")}
          />
          <Input
            label="Email"
            type="email"
            autoComplete="email"
            error={form.formState.errors.email?.message}
            {...form.register("email")}
          />
          <Input
            label="Password"
            type="password"
            autoComplete="new-password"
            hint="At least 8 characters"
            error={form.formState.errors.password?.message}
            {...form.register("password")}
          />
          <Input
            label="Phone (optional)"
            type="tel"
            autoComplete="tel"
            error={form.formState.errors.phone?.message}
            {...form.register("phone")}
          />
          <Button type="submit" loading={loading} className="w-full">
            Create account
          </Button>
        </form>

        <p className="mt-6 text-center text-sm text-ink500">
          Already have an account?{" "}
          <Link to="/login" className="text-ink underline underline-offset-2">
            Sign in
          </Link>
        </p>
      </Card>
    </div>
  );
}
