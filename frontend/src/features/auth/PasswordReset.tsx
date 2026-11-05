import { zodResolver } from "@hookform/resolvers/zod";
import { useState } from "react";
import { useForm } from "react-hook-form";
import { Link, useNavigate, useSearchParams } from "react-router-dom";
import { z } from "zod";
import { AuthField, AuthFieldGroup } from "./AuthField";
import { AuthButton, AuthError, AuthLayout } from "./AuthLayout";
import { useToast } from "../../components/useToast";
import {
  extractAuthErrorMessage,
  forgotPassword,
  lookupOrgs,
  resetPassword,
  type OrgSummary,
} from "../../lib/api/authClient";

const forgotEmailSchema = z.object({
  email: z.string().email("Enter a valid email address"),
});

const resetSchema = z
  .object({
    reset_token: z.string().min(1, "Reset token is required"),
    new_password: z.string().min(8, "Password must be at least 8 characters"),
    confirm_password: z.string().min(1, "Confirm your password"),
  })
  .refine((values) => values.new_password === values.confirm_password, {
    message: "Passwords do not match",
    path: ["confirm_password"],
  });

type ForgotEmailForm = z.infer<typeof forgotEmailSchema>;
type ResetForm = z.infer<typeof resetSchema>;

type ForgotStep = "email" | "org";

export default function PasswordReset() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const { showSuccess } = useToast();

  const initialToken = searchParams.get("token") ?? "";
  const mode = initialToken ? "reset" : "forgot";

  const [forgotStep, setForgotStep] = useState<ForgotStep>("email");
  const [email, setEmail] = useState("");
  const [orgs, setOrgs] = useState<OrgSummary[]>([]);
  const [selectedOrgId, setSelectedOrgId] = useState("");
  const [apiError, setApiError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  const forgotForm = useForm<ForgotEmailForm>({
    resolver: zodResolver(forgotEmailSchema),
    defaultValues: { email: "" },
  });

  const resetForm = useForm<ResetForm>({
    resolver: zodResolver(resetSchema),
    defaultValues: {
      reset_token: initialToken,
      new_password: "",
      confirm_password: "",
    },
  });

  async function handleForgotEmailSubmit(values: ForgotEmailForm) {
    setLoading(true);
    setApiError(null);
    try {
      const { orgs: foundOrgs } = await lookupOrgs(values.email);
      setEmail(values.email);

      if (foundOrgs.length === 0) {
        showSuccess(
          "If an account exists for this email, reset instructions have been sent."
        );
        return;
      }

      if (foundOrgs.length === 1) {
        await submitForgotPassword(values.email, foundOrgs[0].org_id);
        return;
      }

      setOrgs(foundOrgs);
      setSelectedOrgId(foundOrgs[0].org_id);
      setForgotStep("org");
    } catch (err) {
      setApiError(extractAuthErrorMessage(err));
    } finally {
      setLoading(false);
    }
  }

  async function submitForgotPassword(userEmail: string, orgId: string) {
    setLoading(true);
    setApiError(null);
    try {
      const { reset_token } = await forgotPassword({
        email: userEmail,
        org_id: orgId,
      });
      showSuccess("Reset email sent. Use your reset token to set a new password.");
      navigate(`/reset-password?token=${encodeURIComponent(reset_token)}`);
    } catch (err) {
      setApiError(extractAuthErrorMessage(err));
    } finally {
      setLoading(false);
    }
  }

  async function handleOrgForgotSubmit() {
    if (!selectedOrgId) {
      setApiError("Select an organization to continue.");
      return;
    }
    await submitForgotPassword(email, selectedOrgId);
  }

  async function handleResetSubmit(values: ResetForm) {
    setLoading(true);
    setApiError(null);
    try {
      await resetPassword({
        reset_token: values.reset_token.trim(),
        new_password: values.new_password,
      });
      showSuccess("Password reset successful. You can sign in now.");
      navigate("/login");
    } catch (err) {
      setApiError(extractAuthErrorMessage(err));
    } finally {
      setLoading(false);
    }
  }

  if (mode === "reset") {
    return (
      <AuthLayout
        title="Reset password"
        subtitle="Enter your reset token and new password"
        footer={
          <p>
            <Link
              to="/login"
              className="text-ochre-light underline-offset-2 hover:underline"
            >
              Back to sign in
            </Link>
          </p>
        }
      >
        {apiError && <AuthError message={apiError} />}

        <form
          className="flex flex-col gap-5"
          onSubmit={resetForm.handleSubmit(handleResetSubmit)}
          noValidate
        >
          <AuthFieldGroup>
            <AuthField
              label="Reset token"
              autoComplete="off"
              placeholder="Reset token"
              error={resetForm.formState.errors.reset_token?.message}
              {...resetForm.register("reset_token")}
            />
            <AuthField
              label="New password"
              type="password"
              autoComplete="new-password"
              placeholder="New password"
              passwordToggle
              hint="At least 8 characters"
              error={resetForm.formState.errors.new_password?.message}
              {...resetForm.register("new_password")}
            />
            <AuthField
              label="Confirm"
              type="password"
              autoComplete="new-password"
              placeholder="Confirm password"
              passwordToggle
              error={resetForm.formState.errors.confirm_password?.message}
              {...resetForm.register("confirm_password")}
            />
          </AuthFieldGroup>
          <AuthButton type="submit" loading={loading} className="w-full">
            Reset password
          </AuthButton>
        </form>
      </AuthLayout>
    );
  }

  return (
    <AuthLayout
      title="Forgot password"
      subtitle={
        forgotStep === "email"
          ? "Enter your email to receive a reset token"
          : "Choose your organization"
      }
      footer={
        <>
          <p>
            Remember your password?{" "}
            <Link
              to="/login"
              className="font-semibold text-ochre-light underline-offset-2 hover:underline"
            >
              Sign in
            </Link>
          </p>
          <p className="mt-2">
            <Link
              to="/reset-password"
              className="text-ochre-light underline-offset-2 hover:underline"
            >
              Already have a token?
            </Link>
          </p>
        </>
      }
    >
      {apiError && <AuthError message={apiError} />}

      {forgotStep === "email" && (
        <form
          className="flex flex-col gap-5"
          onSubmit={forgotForm.handleSubmit(handleForgotEmailSubmit)}
          noValidate
        >
          <AuthFieldGroup>
            <AuthField
              label="Email"
              type="email"
              autoComplete="email"
              placeholder="you@agency.et"
              error={forgotForm.formState.errors.email?.message}
              {...forgotForm.register("email")}
            />
          </AuthFieldGroup>
          <AuthButton type="submit" loading={loading} className="w-full">
            Send reset token
          </AuthButton>
        </form>
      )}

      {forgotStep === "org" && (
        <div className="flex flex-col gap-5">
          <fieldset>
            <legend className="sr-only">Organization</legend>
            <AuthFieldGroup>
              {orgs.map((org) => (
                <label
                  key={org.org_id}
                  className={[
                    "flex cursor-pointer items-center gap-3 px-4 py-3 sm:px-5",
                    selectedOrgId === org.org_id
                      ? "bg-ochre/10"
                      : "hover:bg-white/[0.03]",
                  ].join(" ")}
                >
                  <input
                    type="radio"
                    name="org"
                    value={org.org_id}
                    checked={selectedOrgId === org.org_id}
                    onChange={() => setSelectedOrgId(org.org_id)}
                    className="accent-ochre"
                  />
                  <span
                    className={[
                      "text-base",
                      selectedOrgId === org.org_id
                        ? "font-medium text-paper"
                        : "text-auth-muted",
                    ].join(" ")}
                  >
                    {org.org_name}
                  </span>
                </label>
              ))}
            </AuthFieldGroup>
          </fieldset>
          <div className="flex gap-3">
            <AuthButton
              type="button"
              variant="secondary"
              onClick={() => {
                setForgotStep("email");
                setApiError(null);
              }}
              className="flex-1"
            >
              Back
            </AuthButton>
            <AuthButton
              type="button"
              loading={loading}
              onClick={handleOrgForgotSubmit}
              className="flex-1"
            >
              Send reset token
            </AuthButton>
          </div>
        </div>
      )}
    </AuthLayout>
  );
}
