package rabbit

const (
	EventPumpSourceToSupplier    = "EventPumpSourceToSupplier"
	EventPumpSupplierToSource    = "EventPumpSupplierToSource"
	EventPumpExpeditorToSupplier = "EventPumpExpeditorToSupplier"

	EventLeopartPayClient      = "EventLeopartPayClient"
	EventLeopartUpdateQuantity = "EventLeopartUpdateQuantity"
	EventLeopartSetInWayStatus = "EventLeopartSetInWayStatus"
)

const (
	QueuePumpSourceToSupplier    = "queue_pump_source_to_supplier"
	QueuePumpSupplierToSource    = "queue_pump_supplier_to_source"
	QueuePumpExpeditorToSupplier = "queue_pump_expeditor_to_supplier"

	QueueLeopartPayClient      = "queue_leopart_pay_client"
	QueueLeopartUpdateQuantity = "queue_leopart_update_quantity"
	QueueLeopartSetInWayStatus = "queue_leopart_set_in_way_status"
)
