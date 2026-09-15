package dev.ricardopena.deskpanel;

import android.os.Bundle;
import android.view.WindowManager;

import com.getcapacitor.BridgeActivity;

/**
 * Ponto de entrada nativo do DeskPanel. Duas responsabilidades adicionadas
 * à casca gerada pelo Capacitor (PROJECT.md §10.1):
 *
 *  - registrar o plugin local SecureTokenStoragePlugin (token no Android
 *    Keystore, nunca em Preferences ou arquivo);
 *  - manter a tela ativa enquanto o painel estiver em primeiro plano
 *    ("keep awake"), já que este é um dispositivo dedicado ao painel.
 *
 * Não verificado por build real neste ambiente (sem Android SDK/Gradle) —
 * ver docs/STATUS.md.
 */
public class MainActivity extends BridgeActivity {

    {
        registerPlugin(SecureTokenStoragePlugin.class);
    }

    @Override
    public void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        getWindow().addFlags(WindowManager.LayoutParams.FLAG_KEEP_SCREEN_ON);
    }
}
