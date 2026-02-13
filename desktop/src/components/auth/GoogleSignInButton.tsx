import { useEffect, useRef, useCallback } from "react";
import { useNavigate } from "react-router-dom";
import { useAuthStore } from "../../stores/authStore";

const GOOGLE_CLIENT_ID = import.meta.env.VITE_GOOGLE_CLIENT_ID as string | undefined;

declare global {
  interface Window {
    google?: {
      accounts: {
        id: {
          initialize: (config: {
            client_id: string;
            callback: (response: { credential: string }) => void;
            auto_select?: boolean;
          }) => void;
          renderButton: (
            element: HTMLElement,
            config: {
              theme?: "outline" | "filled_blue" | "filled_black";
              size?: "large" | "medium" | "small";
              width?: number;
              text?: "signin_with" | "signup_with" | "continue_with" | "signin";
              shape?: "rectangular" | "pill" | "circle" | "square";
            }
          ) => void;
        };
      };
    };
  }
}

export default function GoogleSignInButton() {
  const navigate = useNavigate();
  const { loginWithGoogle, clearError } = useAuthStore();
  const buttonRef = useRef<HTMLDivElement>(null);
  const initializedRef = useRef(false);

  const handleCredentialResponse = useCallback(
    async (response: { credential: string }) => {
      clearError();
      try {
        await loginWithGoogle(response.credential);
        navigate("/");
      } catch {
        // error is set in store
      }
    },
    [loginWithGoogle, navigate, clearError]
  );

  useEffect(() => {
    if (!GOOGLE_CLIENT_ID || initializedRef.current) return;

    const initializeGoogle = () => {
      if (!window.google || !buttonRef.current) return;
      initializedRef.current = true;

      window.google.accounts.id.initialize({
        client_id: GOOGLE_CLIENT_ID,
        callback: handleCredentialResponse,
      });

      window.google.accounts.id.renderButton(buttonRef.current, {
        theme: "outline",
        size: "large",
        width: 320,
        text: "signin_with",
        shape: "rectangular",
      });
    };

    if (window.google) {
      initializeGoogle();
      return;
    }

    const script = document.createElement("script");
    script.src = "https://accounts.google.com/gsi/client";
    script.async = true;
    script.defer = true;
    script.onload = initializeGoogle;
    document.head.appendChild(script);

    return () => {
      if (script.parentNode) {
        script.parentNode.removeChild(script);
      }
    };
  }, [handleCredentialResponse]);

  if (!GOOGLE_CLIENT_ID) return null;

  return (
    <>
      <div ref={buttonRef} className="flex justify-center" />

      <div className="my-4 flex items-center gap-3">
        <div className="h-px flex-1 bg-gray-300 dark:bg-gray-600" />
        <span className="text-xs text-gray-500 dark:text-gray-400">or</span>
        <div className="h-px flex-1 bg-gray-300 dark:bg-gray-600" />
      </div>
    </>
  );
}
