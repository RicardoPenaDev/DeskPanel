package dev.ricardopena.deskpanel;

import android.os.Bundle;
import android.view.WindowManager;

import com.getcapacitor.BridgeActivity;

/**
 * Ponto de entrada nativo do DeskPanel. Responsabilidades adicionadas
 * à casca gerada pelo Capacitor (PROJECT.md §10.1):
 *
 *  - registrar o plugin local SecureTokenStoragePlugin (token no Android
 *    Keystore, nunca em Preferences ou arquivo);
 *  - registrar o plugin local DeviceControlPlugin (keep awake, modo
 *    imersivo e brilho reduzido, todos ajustáveis pela tela de
 *    Configurações — PROJECT.md §10.2-D);
 *  - manter a tela ativa por padrão assim que a Activity é criada ("keep
 *    awake"), antes mesmo do JS carregar a preferência salva — a tela de
 *    Configurações pode desligar isso depois via DeviceControlPlugin.
 *
 * Não verificado por build real neste ambiente (sem Android SDK/Gradle) —
 * ver docs/STATUS.md.
 */
public class MainActivity extends BridgeActivity {

    {
        registerPlugin(SecureTokenStoragePlugin.class);
        registerPlugin(DeviceControlPlugin.class);
    }

    @Override
    public void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        getWindow().addFlags(WindowManager.LayoutParams.FLAG_KEEP_SCREEN_ON);
    }
}
