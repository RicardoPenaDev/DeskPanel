package dev.ricardopena.deskpanel;

import android.app.Activity;
import android.os.Build;
import android.view.View;
import android.view.WindowInsets;
import android.view.WindowInsetsController;
import android.view.WindowManager;

import com.getcapacitor.Plugin;
import com.getcapacitor.PluginCall;
import com.getcapacitor.PluginMethod;
import com.getcapacitor.annotation.CapacitorPlugin;

/**
 * Controles nativos da tela do painel que não têm plugin oficial do
 * Capacitor 6 e não precisam de uma dependência externa para tão pouco
 * código (PROJECT.md §10.2-D — tela de Configurações): manter a tela
 * ligada, modo imersivo (esconder barras de sistema) e brilho reduzido.
 * Nenhum dos três pede permissão especial do sistema: "keep awake" e
 * "imersivo" são flags da própria janela, e o brilho aqui é um atributo da
 * janela do app (WindowManager.LayoutParams.screenBrightness), não do
 * sistema — não precisa de android.permission.WRITE_SETTINGS.
 *
 * Não verificado por build real neste ambiente (sem Android SDK/Gradle) —
 * ver docs/STATUS.md. Mesma ressalva já registrada para
 * SecureTokenStoragePlugin (ADR-0004).
 */
@CapacitorPlugin(name = "DeviceControl")
public class DeviceControlPlugin extends Plugin {

    @PluginMethod
    public void setKeepAwake(PluginCall call) {
        boolean enabled = call.getBoolean("enabled", true);
        Activity activity = getActivity();
        if (activity == null) {
            call.reject("activity indisponível");
            return;
        }
        activity.runOnUiThread(() -> {
            if (enabled) {
                activity.getWindow().addFlags(WindowManager.LayoutParams.FLAG_KEEP_SCREEN_ON);
            } else {
                activity.getWindow().clearFlags(WindowManager.LayoutParams.FLAG_KEEP_SCREEN_ON);
            }
        });
        call.resolve();
    }

    @PluginMethod
    public void setImmersiveMode(PluginCall call) {
        boolean enabled = call.getBoolean("enabled", false);
        Activity activity = getActivity();
        if (activity == null) {
            call.reject("activity indisponível");
            return;
        }
        activity.runOnUiThread(() -> applyImmersiveMode(activity, enabled));
        call.resolve();
    }

    @PluginMethod
    public void setDimBrightness(PluginCall call) {
        boolean enabled = call.getBoolean("enabled", false);
        Activity activity = getActivity();
        if (activity == null) {
            call.reject("activity indisponível");
            return;
        }
        activity.runOnUiThread(() -> {
            WindowManager.LayoutParams params = activity.getWindow().getAttributes();
            // 0.05f é o mínimo utilizável sem apagar a tela de vez;
            // BRIGHTNESS_OVERRIDE_NONE devolve o controle ao brilho do
            // sistema.
            params.screenBrightness = enabled ? 0.05f : WindowManager.LayoutParams.BRIGHTNESS_OVERRIDE_NONE;
            activity.getWindow().setAttributes(params);
        });
        call.resolve();
    }

    @SuppressWarnings("deprecation")
    private void applyImmersiveMode(Activity activity, boolean enabled) {
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.R) {
            WindowInsetsController controller = activity.getWindow().getInsetsController();
            if (controller == null) return;
            if (enabled) {
                controller.hide(WindowInsets.Type.systemBars());
                controller.setSystemBarsBehavior(WindowInsetsController.BEHAVIOR_SHOW_TRANSIENT_BARS_BY_SWIPE);
            } else {
                controller.show(WindowInsets.Type.systemBars());
            }
            return;
        }

        // Fallback para Android < 11 (API 30): flags legadas de
        // SystemUiVisibility, deprecadas mas ainda funcionais.
        View decorView = activity.getWindow().getDecorView();
        if (enabled) {
            decorView.setSystemUiVisibility(
                View.SYSTEM_UI_FLAG_LAYOUT_STABLE
                    | View.SYSTEM_UI_FLAG_LAYOUT_HIDE_NAVIGATION
                    | View.SYSTEM_UI_FLAG_LAYOUT_FULLSCREEN
                    | View.SYSTEM_UI_FLAG_HIDE_NAVIGATION
                    | View.SYSTEM_UI_FLAG_FULLSCREEN
                    | View.SYSTEM_UI_FLAG_IMMERSIVE_STICKY
            );
        } else {
            decorView.setSystemUiVisibility(View.SYSTEM_UI_FLAG_VISIBLE);
        }
    }
}
