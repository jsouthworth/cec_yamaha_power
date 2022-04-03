#include <Arduino.h>

#define IR_SEND_PIN         3
#define I2C_BUS_ADDR 0x08

#include <Wire.h>
#include <IRremote.h>

typedef enum {
	IR_PROTO_NEC = 1,
	IR_PROTO_ONKYO
} i2c_ir_proto;

class IRReq {
public:
	struct data {
		uint8_t proto = 0;
		uint16_t addr = 0;
		uint16_t cmd = 0;
	};

	IRReq() {};

	bool get(data* got) {
		bool out = swap();
		*got = *readable;
		return out;
	}

	bool get_ptr(const data** got) {
		bool out = swap();
		*got = readable;
		return out;
	}

	void update(data* data) {
		*writeable = *data;
		updated = true;
	}

private:
	bool swap() {
		noInterrupts();
		bool out = updated;
		if (updated) {
			data* tmp = readable;
			readable = writeable;
			writeable = tmp;
			updated = false;
		}
		interrupts();
		return out;
	}

	data buf[2];
	data* volatile readable = &buf[0];
	data* volatile writeable = &buf[1];
	volatile bool updated = false;
};

IRReq req;

void i2c_receive(int n) {
	Serial.print("i2c_receive ");
	Serial.println(n);
	if (n != 5)
		return;

	IRReq::data data;
	data.proto = Wire.read();
	data.addr = Wire.read();
	data.addr = data.addr << 8;
    data.addr |= Wire.read();
	data.cmd = Wire.read();
	data.cmd = data.cmd << 8;
    data.cmd |= Wire.read();

	req.update(&data);
}

void setup() {
	Serial.begin(115200);
	IrSender.begin(IR_SEND_PIN, ENABLE_LED_FEEDBACK);
	Serial.print("NEC: ");
	Serial.println(IR_PROTO_NEC);
	Serial.print("ONKYO: ");
	Serial.println(IR_PROTO_ONKYO);
	Wire.begin(I2C_BUS_ADDR);
	Wire.onReceive(i2c_receive);
	Serial.print(F("Listening on i2c bus addr "));
	Serial.println(I2C_BUS_ADDR, HEX);
	Serial.print(F("Ready to send IR signals at pin "));
    Serial.println(IR_SEND_PIN);
}

void ir_send(const IRReq::data* req) {
	Serial.print("action ");
	Serial.print(req->proto, HEX);
	Serial.print("   ");
    Serial.print(req->addr, HEX);
    Serial.print("   ");
    Serial.print(req->cmd, HEX);
    Serial.println();
	switch(req->proto) {
	case IR_PROTO_NEC:
		IrSender.sendNEC(req->addr, req->cmd, 1);
		break;
	case IR_PROTO_ONKYO:
		IrSender.sendOnkyo(req->addr, req->cmd, 1);
		break;
    }
}

void loop() {
	const IRReq::data* data;
	if (req.get_ptr(&data))
		ir_send(data);
}
