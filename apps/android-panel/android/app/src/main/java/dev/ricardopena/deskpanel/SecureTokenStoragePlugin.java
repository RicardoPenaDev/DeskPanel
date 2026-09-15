package dev.ricardopena.deskpanel;

import android.content.SharedPreferences;
import android.util.Log;

import androidx.security.crypto.EncryptedSharedPreferences;
import androidx.security.crypto.MasterKey;

import com.getcapacitor.JSObject;
import com.getcapacitor.Plugin;
import com.getcapacitor.PluginCall;
import com.getcapacitor.PluginMethod;
import com.getcapacitor.annotation.CapacitorPlugin;

import java.io.IOException;
import java.security.GeneralSecurityException;

/**
 * Plugin Capacitor local (não publicado como pacote separado) que guarda o
 * token de acesso do DeskPanel no Android Keystore por meio de
 * EncryptedSharedPreferences (PROJECT.md §10.3, §12.8, docs/SECURITY.md).
 *
 * Decisão de projeto: escrito à mão em vez de usar um plugin de terceiros
 * não auditado, para manter o armazenamento do token sob controle direto —
 * ver docs/DECISIONS.md ADR-0004. Este código não pôde ser compilado nem
 * testado neste ambiente (sandbox sem Android SDK/Gradle); ver
 * docs/STATUS.md para o aviso de verificação pendente.
 *
 * Nunca registra o valor do token em log, mesmo em caso de erro.
 */
@CapacitorPlugin(name = "SecureTokenStorage")
public class SecureTokenStoragePlugin extends Plugin {

    private static final String TAG = "SecureTokenStorage";
    private static final String PREFS_FILE_NAME = "deskpanel_secure_tokens";
    private static final String KEY_PREFIX = "token:";

    private SharedPreferences encryptedPrefs() throws GeneralSecurityException, IOException {
        MasterKey masterKey = new MasterKey.Builder(getContext())
                .setKeyScheme(MasterKey.KeyScheme.AES256_GCM)
                .build();

        return EncryptedSharedPreferences.create(
                getContext(),
                PREFS_FILE_NAME,
                masterKey,
                EncryptedSharedPreferences.PrefKeyEncryptionScheme.AES256_SIV,
                EncryptedSharedPreferences.PrefValueEncryptionScheme.AES256_GCM
        );
    }

    @PluginMethod
    public void setToken(PluginCall call) {
        String deviceId = call.getString("deviceId");
        String token = call.getString("token");

        if (deviceId == null || deviceId.isEmpty() || token == null || token.isEmpty()) {
            call.reject("deviceId e token são obrigatórios");
            return;
        }

        try {
            encryptedPrefs().edit().putString(KEY_PREFIX + deviceId, token).apply();
            call.resolve();
        } catch (GeneralSecurityException | IOException e) {
            Log.e(TAG, "falha ao gravar token no Keystore: " + e.getClass().getSimpleName());
            call.reject("Não foi possível salvar o token com segurança neste dispositivo");
        }
    }

    @PluginMethod
    public void getToken(PluginCall call) {
        String deviceId = call.getString("deviceId");
        if (deviceId == null || deviceId.isEmpty()) {
            call.reject("deviceId é obrigatório");
            return;
        }

        try {
            String token = encryptedPrefs().getString(KEY_PREFIX + deviceId, null);
            JSObject result = new JSObject();
            result.put("token", token);
            call.resolve(result);
        } catch (GeneralSecurityException | IOException e) {
            Log.e(TAG, "falha ao ler token do Keystore: " + e.getClass().getSimpleName());
            call.reject("Não foi possível ler o token salvo neste dispositivo");
        }
    }

    @PluginMethod
    public void clearToken(PluginCall call) {
        String deviceId = call.getString("deviceId");
        if (deviceId == null || deviceId.isEmpty()) {
            call.reject("deviceId é obrigatório");
            return;
        }

        try {
            encryptedPrefs().edit().remove(KEY_PREFIX + deviceId).apply();
            call.resolve();
        } catch (GeneralSecurityException | IOException e) {
            Log.e(TAG, "falha ao remover token do Keystore: " + e.getClass().getSimpleName());
            call.reject("Não foi possível remover o token salvo neste dispositivo");
        }
    }
}
