# BarCVVR

Code raspberry pour débimètres : 

```python
#!/usr/bin/python3.5

import RPi.GPIO as GPIO
import time, sys
import requests
import asyncio

GPIO.setmode(GPIO.BOARD)
inpt = 8
inpt2 = 10
GPIO.setup(inpt,GPIO.IN)
GPIO.setup(inpt2,GPIO.IN)
rate_cnt = 0.0000
rate_cnt2 = 0.0000
constant = 9230
time_new = 0.0
state = GPIO.input(inpt)
state2 = GPIO.input(inpt2)

loop = asyncio.get_event_loop()  

async def send_data_blonde(liter):
    def do_req():
        requests.put("http://[URL]/beerflows/1", data={'_method': 'patch', 'beerflow[quantity]': liter, 'beerflow[drink_id]': '1'})
    loop.run_in_executor(None, do_req)
    
async def send_data_special(liter):
    def do_req():
        requests.put("http://[URL]/beerflows/2", data={'_method': 'patch', 'beerflow[quantity]': liter, 'beerflow[drink_id]': '9'})
    loop.run_in_executor(None, do_req)

while True:
    time_new = time.time() + 1
    rate_cnt = 0
    rate_cnt2 = 0
    while time.time() <= time_new:
        if GPIO.input(inpt)!= state:
            rate_cnt += 1
            state = GPIO.input(inpt)
        if GPIO.input(inpt2)!= state2:
            rate_cnt2 += 1
            state2 = GPIO.input(inpt2)
    liter = round(rate_cnt / constant,4)
    liter2 = round(rate_cnt2 / constant,4)
    if liter > 0:
      loop.run_until_complete(send_data_blonde(liter))
    if liter2 > 0:
      loop.run_until_complete(send_data_special(liter2)) 

loop.close()
GPIO.cleanup()
```


ESP32  
```c++
#include <WiFi.h>
#include <HTTPClient.h>

// ==========================================
// CONFIGURATION RÉSEAU ET SERVEUR
// ==========================================
const char* ssid = "Velivolistes42";
const char* password = "0477668108";
// L'URL de l'API Go. L'utilisation de http:// évite les problèmes de certificats 
// SSL/TLS avec le client HTTP basique de l'ESP32.
const char* serverUrlBase = "http://bar.cvvr.fr";

// ==========================================
// CONFIGURATION DES DÉBITMÈTRES
// ==========================================
// Broches (Pins) recevant le signal des turbines (débitmètres).
const int SENSOR_PIN_BLONDE = 22;
const int SENSOR_PIN_SPECIAL = 23;

// Constante d'étalonnage : Nombre d'impulsions par litre de liquide.
// Ajuster si la quantité annoncée diffère de la quantité réele (ex: calibrage au pichet).
const float CALIBRATION_CONSTANT = 9230.0;

// Compteurs bruts d'impulsions. Le mot-clé 'volatile' indique au compilateur
// que ces variables peuvent changer à tout moment via des interruptions matérielles.
volatile unsigned long pulseCounterBlonde = 0;
volatile unsigned long pulseCounterSpecial = 0;

// ==========================================
// GESTION ANTI-REBOND (DEBOUNCE)
// ==========================================
// Les turbines créent un signal électrique bruité (plusieurs impulsions fantômes pour un seul tour).
// Ces variables permettent d'ignorer ces parasites.
volatile unsigned long lastMicrosBlonde = 0;
volatile unsigned long lastMicrosSpecial = 0;

// Temps minimum entre 2 vraies impulsions en microsecondes.
// 50µs = Max 20 000 impulsions par seconde.
const unsigned long debounceMicros = 50; 

// ==========================================
// TAMPON (BUFFER) ET INTERVALLE RESEAU
// ==========================================
unsigned long lastSendTime = 0;
// Intervalle de cycle principal : toutes les 2 secondes on traite les débits.
const long sendInterval = 2000;

// Si la connexion WiFi coupe ou si le serveur ne répond pas, ces variables
// accumulent les volumes pour éviter de les perdre entre 2 requêtes.
float pendingLitersBlonde = 0.0;
float pendingLitersSpecial = 0.0;

// ==========================================
// INTERRUPTIONS MATÉRIELLES (ISR)
// ==========================================
// Ces fonctions s'exécutent de façon prioritaire lorsqu'un signal électrique change.
// IRAM_ATTR force la fonction en RAM (et non en Flash CPU) pour une vitesse d'exécution optimale.

void IRAM_ATTR onPulseBlonde() {
  unsigned long currentMicros = micros();
  // Ne compte l'impulsion que si la précédente a eu lieu il y a plus de 'debounceMicros'
  if (currentMicros - lastMicrosBlonde > debounceMicros) {
    pulseCounterBlonde++;
    lastMicrosBlonde = currentMicros;
  }
}

void IRAM_ATTR onPulseSpecial() {
  unsigned long currentMicros = micros();
  if (currentMicros - lastMicrosSpecial > debounceMicros) {
    pulseCounterSpecial++;
    lastMicrosSpecial = currentMicros;
  }
}

// ==========================================
// ROUTINES DE CONNEXION ET TRANSFERT
// ==========================================

// Initialise la connexion WiFi et bloque tant qu'elle n'est pas établie
void setup_wifi() {
  delay(10);
  Serial.println();
  Serial.print("Connexion au WiFi : ");
  Serial.println(ssid);
  
  WiFi.begin(ssid, password);
  
  while (WiFi.status() != WL_CONNECTED) {
    delay(500);
    Serial.print(".");
  }
  
  Serial.println("\nWiFi connecté !");
  Serial.print("Adresse IP : ");
  Serial.println(WiFi.localIP());
}

// Fonction utilitaire pour envoyer le volume à l'API Rest Go.
bool sendFlowData(const char* flowId, float quantity, const char* drinkId) {
  if (WiFi.status() == WL_CONNECTED) {
    HTTPClient http;
    // Ciblage du endpoint (ex: http://bar.cvvr.fr/beerflows/1)
    String serverPath = String(serverUrlBase) + "/beerflows/" + String(flowId);
    
    http.begin(serverPath.c_str());
    http.addHeader("Content-Type", "application/x-www-form-urlencoded");
    
    // Pour compatibilité API, on intercepte la méthodologie avec "_method=patch"
    // Et le paramètre natif beerflow[quantity]=X & beerflow[drink_id]=Y
    String postData = "_method=patch&beerflow[quantity]=" + String(quantity, 4) + "&beerflow[drink_id]=" + String(drinkId);
    int httpResponseCode = http.POST(postData);
    
    bool success = false;
    // Si l'API valide la création, on acquitte.
    if (httpResponseCode == 200 || httpResponseCode == 302) {
      success = true;
    } else {
      Serial.printf("❌ Erreur de transfert HTTP (code %d) pour le capteur %s\n", httpResponseCode, flowId);
    }
    
    http.end();
    return success;
  } else {
    Serial.println("❌ Perte du signal WiFi. Les données de la bière sont sauvegardées en interne.");
    return false;
  }
}

// ==========================================
// DEMARRAGE ET BOUCLE PRINCIPALE
// ==========================================

void setup() {
  Serial.begin(115200);
  
  // Utilise les résistances de pull-up internes pour stabiliser les capteurs quand le signal est ouvert
  pinMode(SENSOR_PIN_BLONDE, INPUT_PULLUP);
  pinMode(SENSOR_PIN_SPECIAL, INPUT_PULLUP);
  
  // Associe la broche électrique à l'interruption logicielle. Déclenchement sur front descendant (FALLING).
  attachInterrupt(digitalPinToInterrupt(SENSOR_PIN_BLONDE), onPulseBlonde, FALLING);
  attachInterrupt(digitalPinToInterrupt(SENSOR_PIN_SPECIAL), onPulseSpecial, FALLING);
  
  setup_wifi();
  lastSendTime = millis();
}

void loop() {
  // L'horloge interne arrive à l'échéance des 'sendInterval' (toutes les 2s)
  if (millis() - lastSendTime >= sendInterval) {
    lastSendTime = millis(); 
    
    unsigned long currentPulsesBlonde;
    unsigned long currentPulsesSpecial;
    
    // DÉBUT DE SECTION CRITIQUE
    // On désactive les interruptions le temps d'une fraction de seconde
    // Pour "photographier" la valeur des compteurs volatils avant de les remettre à 0 sans interférences.
    noInterrupts(); 
    currentPulsesBlonde = pulseCounterBlonde;
    pulseCounterBlonde = 0;
    currentPulsesSpecial = pulseCounterSpecial;
    pulseCounterSpecial = 0;
    interrupts(); 
    // FIN DE SECTION CRITIQUE
    
    // Conversion mathématique des impulsions en Litres (L)
    float litersBlonde = (float)currentPulsesBlonde / CALIBRATION_CONSTANT;
    float litersSpecial = (float)currentPulsesSpecial / CALIBRATION_CONSTANT;
    
    // Stockage dans les tampons en attente d'expédition
    pendingLitersBlonde += litersBlonde;
    pendingLitersSpecial += litersSpecial;
    
    // --- EVALUATION ET ENVOI: BIÈRE BLONDE ---
    // Si des litres s'accumulent (supérieur à une marge d'erreur infinitésimale)
    if (pendingLitersBlonde > 0.0001) {
      Serial.printf("🍺 Détection Blonde : %.4f L\n", pendingLitersBlonde);
      // Tentative d'envoi. Si réussie, on vide le tampon.
      if (sendFlowData("1", pendingLitersBlonde, "1")) {
        Serial.println("✅ Envoyé au serveur avec succès !");
        pendingLitersBlonde = 0; 
      }
    }
    
    // --- EVALUATION ET ENVOI: BIÈRE SPÉCIALE ---
    if (pendingLitersSpecial > 0.0001) {
      Serial.printf("🍻 Détection Spéciale : %.4f L\n", pendingLitersSpecial);
      if (sendFlowData("2", pendingLitersSpecial, "9")) {
        Serial.println("✅ Envoyé au serveur avec succès !");
        pendingLitersSpecial = 0; 
      }
    }
  }
}
```
